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

type fakeUserProfileService struct {
	err bool
}

func (f *fakeUserProfileService) GetProfile(_ context.Context, _ int32) (service.ProfileOutput, error) {
	if f.err {
		return service.ProfileOutput{}, errors.New("error de servicio")
	}
	return service.ProfileOutput{}, nil
}

func (f *fakeUserProfileService) Create(_ context.Context, input service.CreateProfileInput) (service.ProfileOutput, error) {
	if f.err {
		return service.ProfileOutput{}, errors.New("error de servicio")
	}
	return service.ProfileOutput{
		UserID:     input.UserID,
		Name:       input.Name,
		DepthLevel: input.DepthLevel,
	}, nil
}

func (f *fakeUserProfileService) UpdateProfile(_ context.Context, _ service.UpdateProfileInput) (service.ProfileOutput, error) {
	if f.err {
		return service.ProfileOutput{}, errors.New("error de servicio")
	}
	return service.ProfileOutput{}, nil
}

func (f *fakeUserProfileService) AddInterest(_ context.Context, _ int32, _ int32) (service.ProfileOutput, error) {
	if f.err {
		return service.ProfileOutput{}, errors.New("error de servicio")
	}
	return service.ProfileOutput{}, nil
}

func (f *fakeUserProfileService) RemoveInterest(_ context.Context, _ int32, _ int32) (service.ProfileOutput, error) {
	if f.err {
		return service.ProfileOutput{}, errors.New("error de servicio")
	}
	return service.ProfileOutput{}, nil
}

func (f *fakeUserProfileService) AddFocusType(_ context.Context, _ int32, _ int32) (service.ProfileOutput, error) {
	if f.err {
		return service.ProfileOutput{}, errors.New("error de servicio")
	}
	return service.ProfileOutput{}, nil
}

func (f *fakeUserProfileService) RemoveFocusType(_ context.Context, _ int32, _ int32) (service.ProfileOutput, error) {
	if f.err {
		return service.ProfileOutput{}, errors.New("error de servicio")
	}
	return service.ProfileOutput{}, nil
}

func newUserProfileTestRouter(svc transport.UserProfileService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := transport.NewUserProfileHandler(svc)
	r := gin.New()
	r.POST("/api/v1/users", h.CreateProfile)
	return r
}

func doCreateProfile(r *gin.Engine, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

const validProfileBody = `{
	"user_id":       1,
	"name":          "Yerald Sánchez",
	"depth_level":   "intermedio",
	"focus_ids":     [1, 2],
	"interests_ids": [3, 4]
}`

func TestCreateProfile_Success(t *testing.T) {
	r := newUserProfileTestRouter(&fakeUserProfileService{})
	w := doCreateProfile(r, validProfileBody)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d — body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Data struct {
			Profile struct {
				UserID int32  `json:"user_id"`
				Name   string `json:"name"`
			} `json:"profile"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if resp.Data.Profile.UserID == 0 {
		t.Fatal("expected user_id in response, got 0")
	}
	if resp.Data.Profile.Name != "Yerald Sánchez" {
		t.Fatalf("expected name 'Yerald Sánchez', got %q", resp.Data.Profile.Name)
	}
}

func TestCreateProfile_ServiceError(t *testing.T) {
	r := newUserProfileTestRouter(&fakeUserProfileService{err: true})
	w := doCreateProfile(r, validProfileBody)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestCreateProfile_CamposFaltantes(t *testing.T) {
	r := newUserProfileTestRouter(&fakeUserProfileService{})

	cases := []struct {
		name string
		body string
	}{
		{"sin user_id", `{"name":"Luis","depth_level":"intermedio","focus_ids":[1],"interests_ids":[2]}`},
		{"sin name", `{"user_id":1,"depth_level":"intermedio","focus_ids":[1],"interests_ids":[2]}`},
		{"sin depth_level", `{"user_id":1,"name":"Luis","focus_ids":[1],"interests_ids":[2]}`},
		{"sin focus_ids", `{"user_id":1,"name":"Luis","depth_level":"intermedio","interests_ids":[2]}`},
		{"sin interests_ids", `{"user_id":1,"name":"Luis","depth_level":"intermedio","focus_ids":[1]}`},
		{"body vacío", `{}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := doCreateProfile(r, tc.body)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d — body: %s", w.Code, w.Body.String())
			}
		})
	}
}
