package service

import (
	"api-cultura-conecta/internal/apperrors"
	db "api-cultura-conecta/internal/db/generated"
	"context"
	"regexp"
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
	q         db.Querier
	jwtSecret []byte
	tokenTTL  time.Duration
}

func NewAuthService(q db.Querier, jwtSecret string) *AuthService {
	return &AuthService{
		q:         q,
		jwtSecret: []byte(jwtSecret),
		tokenTTL:  24 * time.Hour,
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

func (s *AuthService) Login(ctx context.Context, input LoginInput) (string, error) {
	user, err := s.q.GetUserByEmail(ctx, input.Email)
	if err != nil {
		return "", apperrors.ErrInvalidCredentials
	}
	if user.PasswordHash == "" {
		return "", apperrors.ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return "", apperrors.ErrInvalidCredentials
	}
	return s.generateToken(user.ID)
}

func (s *AuthService) generateToken(userID int32) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": now.Add(s.tokenTTL).Unix(),
		"iat": now.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}
