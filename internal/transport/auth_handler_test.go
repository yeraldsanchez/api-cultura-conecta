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
	users map[string]string // email → passwordHash
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

func (f *fakeAuthService) Login(_ context.Context, input service.LoginInput) (string, error) {
	pw, ok := f.users[input.Email]
	if !ok || pw != input.Password {
		return "", errors.New("credenciales inválidas")
	}
	return "fake-token", nil
}

func newAuthTestRouter(svc transport.AuthService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := transport.NewAuthHandler(svc)
	r := gin.New()
	r.POST("/api/v1/auth/register", h.Register)
	r.POST("/api/v1/auth/login", h.Login)
	return r
}

func doRegister(r *gin.Engine, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(body))
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
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 on duplicate email, got %d — body: %s", w.Code, w.Body.String())
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
