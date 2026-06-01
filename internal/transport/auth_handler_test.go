package transport_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	db "api-cultura-conecta/internal/db/generated"
	"api-cultura-conecta/internal/service"
	"api-cultura-conecta/internal/transport"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// fakeQuerier implementa db.Querier con un mapa en memoria.
type fakeQuerier struct {
	users  map[string]db.User
	nextID int32
}

func newFakeQuerier() *fakeQuerier {
	return &fakeQuerier{users: make(map[string]db.User)}
}

func (f *fakeQuerier) CreateUser(_ context.Context, arg db.CreateUserParams) (int32, error) {
	if _, exists := f.users[arg.Email]; exists {
		return 0, &pgconn.PgError{Code: "23505"}
	}
	f.nextID++
	f.users[arg.Email] = db.User{ID: f.nextID, Email: arg.Email, PasswordHash: arg.PasswordHash}
	return f.nextID, nil
}

func (f *fakeQuerier) GetUserByEmail(_ context.Context, email string) (db.User, error) {
	u, ok := f.users[email]
	if !ok {
		return db.User{}, pgx.ErrNoRows
	}
	return u, nil
}

func (f *fakeQuerier) AssignFocusTypeToUser(_ context.Context, _ db.AssignFocusTypeToUserParams) error {
	return nil
}

func (f *fakeQuerier) CreateFocusType(_ context.Context, _ string) (db.FocusType, error) {
	return db.FocusType{}, nil
}

func (f *fakeQuerier) CreateUserProfile(_ context.Context, _ db.CreateUserProfileParams) (int32, error) {
	return 0, nil
}

func (f *fakeQuerier) GetFocusTypes(_ context.Context) ([]db.FocusType, error) {
	return nil, nil
}

func (f *fakeQuerier) GetUserFocusTypes(_ context.Context, _ int32) ([]db.FocusType, error) {
	return nil, nil
}

func (f *fakeQuerier) GetUserProfileByUserId(_ context.Context, _ int32) (db.GetUserProfileByUserIdRow, error) {
	return db.GetUserProfileByUserIdRow{}, nil
}

func (f *fakeQuerier) AssignInterestToUser(_ context.Context, _ db.AssignInterestToUserParams) error {
	return nil
}

func (f *fakeQuerier) CreateCategory(_ context.Context, _ string) (db.Category, error) {
	return db.Category{}, nil
}

func (f *fakeQuerier) GetCategories(_ context.Context) ([]db.Category, error) {
	return nil, nil
}

func (f *fakeQuerier) GetUserInterests(_ context.Context, _ int32) ([]db.Category, error) {
	return nil, nil
}

func newTestRouter(q db.Querier) *gin.Engine {
	gin.SetMode(gin.TestMode)
	svc := service.NewAuthService(q, "test-secret-key")
	h := transport.NewAuthHandler(svc)
	r := gin.New()
	r.POST("/api/v1/auth/register", h.Register)
	return r
}

func doRegister(r *gin.Engine, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestRegister_Success(t *testing.T) {
	r := newTestRouter(newFakeQuerier())
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
	r := newTestRouter(newFakeQuerier())
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
	r := newTestRouter(newFakeQuerier())

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

func TestRegister_WeakPassword(t *testing.T) {
	r := newTestRouter(newFakeQuerier())

	cases := []struct {
		name string
		body string
	}{
		{"muy corta", `{"email":"user@example.com","password":"Ab1!"}`},
		{"sin número", `{"email":"user@example.com","password":"Password!"}`},
		{"sin letra", `{"email":"user@example.com","password":"12345678!"}`},
		{"sin especial", `{"email":"user@example.com","password":"Password1"}`},
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
