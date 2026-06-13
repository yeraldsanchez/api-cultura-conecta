package service

import (
	"api-cultura-conecta/internal/apperrors"
	db "api-cultura-conecta/internal/db/generated"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	reDigit   = regexp.MustCompile(`[0-9]`)
	reLetter  = regexp.MustCompile(`[a-zA-Z]`)
	reSpecial = regexp.MustCompile(`[^a-zA-Z0-9]`)
)

func validatePassword(p string) error {
	switch {
	case len(p) < 8:
		return apperrors.NewValidationError("la contraseña debe tener al menos 8 caracteres")
	case !reDigit.MatchString(p):
		return apperrors.NewValidationError("la contraseña debe contener al menos un número")
	case !reLetter.MatchString(p):
		return apperrors.NewValidationError("la contraseña debe contener al menos una letra")
	case !reSpecial.MatchString(p):
		return apperrors.NewValidationError("la contraseña debe contener al menos un carácter especial")
	}
	return nil
}

type AuthService struct {
	q               db.Querier
	jwtSecret       []byte
	tokenTTL        time.Duration
	refreshTokenTTL time.Duration
}

func NewAuthService(q db.Querier, jwtSecret string) *AuthService {
	return &AuthService{
		q:               q,
		jwtSecret:       []byte(jwtSecret),
		tokenTTL:        15 * time.Minute,
		refreshTokenTTL: 7 * 24 * time.Hour,
	}
}

type CreateUserInput struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=8"`
}

type LoginInput struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required"`
}

func (s *AuthService) Register(ctx context.Context, input CreateUserInput) (*int32, error) {
	if err := validatePassword(input.Password); err != nil {
		return nil, err
	}

	rawHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	hashStr := string(rawHash)

	id, err := s.q.CreateUser(ctx, db.CreateUserParams{
		Email:        input.Email,
		PasswordHash: hashStr,
	})
	if err != nil {
		return nil, apperrors.FromPgx(err, apperrors.UserConstraints)
	}
	return &id, nil
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (string, string, error) {
	user, err := s.q.GetUserByEmail(ctx, input.Email)
	if err != nil {
		return "", "", apperrors.ErrInvalidCredentials
	}
	if user.PasswordHash == "" {
		return "", "", apperrors.ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return "", "", apperrors.ErrInvalidCredentials
	}

	accessToken, err := s.generateToken(user.ID)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := s.createRefreshToken(ctx, user.ID)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *AuthService) RefreshAccessToken(ctx context.Context, refreshToken string) (string, error) {
	tokenHash := hashToken(refreshToken)
	row, err := s.q.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		return "", apperrors.ErrInvalidCredentials
	}
	return s.generateToken(row.UserID)
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	tokenHash := hashToken(refreshToken)
	return s.q.RevokeRefreshToken(ctx, tokenHash)
}

func (s *AuthService) ValidateAccessToken(tokenStr string) (int32, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, apperrors.ErrUnauthorized
		}
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return 0, apperrors.ErrUnauthorized
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, apperrors.ErrUnauthorized
	}

	sub, err := claims.GetSubject()
	if err != nil || sub == "" {
		return 0, apperrors.ErrUnauthorized
	}

	userID, err := strconv.ParseInt(sub, 10, 32)
	if err != nil {
		return 0, apperrors.ErrUnauthorized
	}

	return int32(userID), nil
}

func (s *AuthService) generateToken(userID int32) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": strconv.Itoa(int(userID)),
		"exp": now.Add(s.tokenTTL).Unix(),
		"iat": now.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *AuthService) createRefreshToken(ctx context.Context, userID int32) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	token := hex.EncodeToString(raw)
	tokenHash := hashToken(token)

	_, err := s.q.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(s.refreshTokenTTL),
	})
	if err != nil {
		return "", err
	}
	return token, nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// ErrUnauthorized is re-exported so middleware can compare without importing apperrors.
var ErrUnauthorized = apperrors.ErrUnauthorized
