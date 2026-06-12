package transport_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"api-cultura-conecta/internal/service"
	"api-cultura-conecta/internal/transport"

	"github.com/gin-gonic/gin"
)

// fakeAuthService implementa transport.AuthService en memoria.
type fakeAuthService struct {
	users map[string]string // email → password
}

func newFakeAuthService() *fakeAuthService {
	return &fakeAuthService{users: make(map[string]string)}
}

func (f *fakeAuthService) Register(_ context.Context, input service.CreateUserInput) (*int32, error) {
	if _, exists := f.users[input.Email]; exists {
		return nil, errors.New("email duplicado")
	}
	f.users[input.Email] = input.Password
	id := int32(len(f.users))
	return &id, nil
}

func (f *fakeAuthService) Login(_ context.Context, input service.LoginInput) (string, string, error) {
	pw, ok := f.users[input.Email]
	if !ok || pw != input.Password {
		return "", "", errors.New("credenciales inválidas")
	}
	return "fake-access-token", "fake-refresh-token", nil
}

func (f *fakeAuthService) RefreshAccessToken(_ context.Context, refreshToken string) (string, error) {
	if refreshToken == "fake-refresh-token" {
		return "new-fake-access-token", nil
	}
	return "", errors.New("refresh token inválido")
}

func (f *fakeAuthService) Logout(_ context.Context, _ string) error {
	return nil
}

func (f *fakeAuthService) ValidateAccessToken(tokenStr string) (int32, error) {
	if tokenStr == "fake-access-token" {
		return 1, nil
	}
	return 0, errors.New("token inválido")
}

func newAuthTestRouter(svc transport.AuthService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := transport.NewAuthHandler(svc)
	r := gin.New()
	r.POST("/api/v1/auth/register", h.Register)
	r.POST("/api/v1/auth/login", h.Login)
	r.POST("/api/v1/auth/refresh", h.RefreshToken)
	r.POST("/api/v1/auth/logout", h.Logout)
	return r
}

func doRegister(r *gin.Engine, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func doLogin(r *gin.Engine, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestRegister_Success(t *testing.T) {
	r := newAuthTestRouter(newFakeAuthService())
	w := doRegister(r, `{"email":"user@example.com","password":"Secure1!"}`)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d — body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Data struct {
			UserID *int32 `json:"user_id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if resp.Data.UserID == nil {
		t.Fatal("expected user_id in response, got nil")
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	r := newAuthTestRouter(newFakeAuthService())
	body := `{"email":"dup@example.com","password":"Secure1!"}`

	if w := doRegister(r, body); w.Code != http.StatusCreated {
		t.Fatalf("first registration should succeed, got %d", w.Code)
	}

	w := doRegister(r, body)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 on duplicate email, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestRegister_InvalidFields(t *testing.T) {
	r := newAuthTestRouter(newFakeAuthService())

	cases := []struct {
		name string
		body string
	}{
		{"email malformado", `{"email":"not-an-email","password":"Secure1!"}`},
		{"sin email", `{"password":"Secure1!"}`},
		{"body vacío", `{}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := doRegister(r, tc.body)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d — body: %s", w.Code, w.Body.String())
			}
		})
	}
}

func TestLogin_ReturnsTokenPair(t *testing.T) {
	svc := newFakeAuthService()
	r := newAuthTestRouter(svc)

	doRegister(r, `{"email":"user@example.com","password":"Secure1!"}`)
	w := doLogin(r, `{"email":"user@example.com","password":"Secure1!"}`)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Data struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
			TokenType    string `json:"token_type"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if resp.Data.AccessToken == "" {
		t.Fatal("expected access_token in response")
	}
	if resp.Data.RefreshToken == "" {
		t.Fatal("expected refresh_token in response")
	}
	if resp.Data.TokenType != "Bearer" {
		t.Fatalf("expected token_type Bearer, got %q", resp.Data.TokenType)
	}
}

func TestRefreshToken_Success(t *testing.T) {
	r := newAuthTestRouter(newFakeAuthService())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh",
		strings.NewReader(`{"refresh_token":"fake-refresh-token"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestLogout_Success(t *testing.T) {
	r := newAuthTestRouter(newFakeAuthService())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout",
		strings.NewReader(`{"refresh_token":"fake-refresh-token"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d — body: %s", w.Code, w.Body.String())
	}
}
