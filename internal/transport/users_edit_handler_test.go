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

// fakeEditService controla qué método falla y captura los inputs recibidos.
type fakeEditService struct {
	updateErr       bool
	addInterestErr  bool
	remInterestErr  bool
	addFocusErr     bool
	remFocusErr     bool
	getProfileErr   bool
	capturedUpdate  service.UpdateProfileInput
	capturedUserID  int32
	capturedItemID  int32
	returnedProfile service.ProfileOutput
}

func (f *fakeEditService) Create(_ context.Context, _ service.CreateProfileInput) (service.ProfileOutput, error) {
	return service.ProfileOutput{}, nil
}

func (f *fakeEditService) GetProfile(_ context.Context, userID int32) (service.ProfileOutput, error) {
	f.capturedUserID = userID
	if f.getProfileErr {
		return service.ProfileOutput{}, errors.New("error de servicio")
	}
	return f.returnedProfile, nil
}

func (f *fakeEditService) UpdateProfile(_ context.Context, input service.UpdateProfileInput) (service.ProfileOutput, error) {
	f.capturedUpdate = input
	if f.updateErr {
		return service.ProfileOutput{}, errors.New("error de servicio")
	}
	return f.returnedProfile, nil
}

func (f *fakeEditService) AddInterest(_ context.Context, userID int32, categoryID int32) (service.ProfileOutput, error) {
	f.capturedUserID = userID
	f.capturedItemID = categoryID
	if f.addInterestErr {
		return service.ProfileOutput{}, errors.New("error de servicio")
	}
	return f.returnedProfile, nil
}

func (f *fakeEditService) RemoveInterest(_ context.Context, userID int32, categoryID int32) (service.ProfileOutput, error) {
	f.capturedUserID = userID
	f.capturedItemID = categoryID
	if f.remInterestErr {
		return service.ProfileOutput{}, errors.New("error de servicio")
	}
	return f.returnedProfile, nil
}

func (f *fakeEditService) AddFocusType(_ context.Context, userID int32, focusTypeID int32) (service.ProfileOutput, error) {
	f.capturedUserID = userID
	f.capturedItemID = focusTypeID
	if f.addFocusErr {
		return service.ProfileOutput{}, errors.New("error de servicio")
	}
	return f.returnedProfile, nil
}

func (f *fakeEditService) RemoveFocusType(_ context.Context, userID int32, focusTypeID int32) (service.ProfileOutput, error) {
	f.capturedUserID = userID
	f.capturedItemID = focusTypeID
	if f.remFocusErr {
		return service.ProfileOutput{}, errors.New("error de servicio")
	}
	return f.returnedProfile, nil
}

func newEditTestRouter(svc transport.UserProfileService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := transport.NewUserProfileHandler(svc)
	r := gin.New()
	r.GET("/api/v1/users/:user_id", h.GetProfile)
	r.PATCH("/api/v1/users/:user_id", h.PatchProfile)
	r.POST("/api/v1/users/:user_id/interests", h.AddInterest)
	r.DELETE("/api/v1/users/:user_id/interests/:category_id", h.RemoveInterest)
	r.POST("/api/v1/users/:user_id/focus-types", h.AddFocusType)
	r.DELETE("/api/v1/users/:user_id/focus-types/:focus_type_id", h.RemoveFocusType)
	return r
}

var sampleProfileOutput = service.ProfileOutput{
	UserID:     1,
	Name:       "Yerald Sánchez",
	Email:      "yerald@example.com",
	ProfileID:  10,
	DepthLevel: "intermedio",
	FocusTypes: []service.FocusTypeOutput{{ID: 1, Name: "Música"}},
	Interests:  []service.InterestOutput{{ID: 2, Name: "Jazz"}},
}

// ─── PATCH /users/:user_id ────────────────────────────────────────────────────

func doPatchProfile(r *gin.Engine, userID string, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/"+userID, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestPatchProfile_SoloNombre(t *testing.T) {
	svc := &fakeEditService{returnedProfile: sampleProfileOutput}
	r := newEditTestRouter(svc)
	w := doPatchProfile(r, "1", `{"name":"Nuevo Nombre"}`)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
	if svc.capturedUpdate.UserID != 1 {
		t.Fatalf("expected user_id 1, got %d", svc.capturedUpdate.UserID)
	}
	if svc.capturedUpdate.Name == nil || *svc.capturedUpdate.Name != "Nuevo Nombre" {
		t.Fatalf("expected name 'Nuevo Nombre', got %v", svc.capturedUpdate.Name)
	}
	if svc.capturedUpdate.DepthLevel != nil {
		t.Fatal("expected depth_level to be nil")
	}
}

func TestPatchProfile_SoloDepthLevel(t *testing.T) {
	svc := &fakeEditService{returnedProfile: sampleProfileOutput}
	r := newEditTestRouter(svc)
	w := doPatchProfile(r, "1", `{"depth_level":"avanzado"}`)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
	if svc.capturedUpdate.DepthLevel == nil || *svc.capturedUpdate.DepthLevel != "avanzado" {
		t.Fatalf("expected depth_level 'avanzado', got %v", svc.capturedUpdate.DepthLevel)
	}
	if svc.capturedUpdate.Name != nil {
		t.Fatal("expected name to be nil")
	}
}

func TestPatchProfile_AmbosCampos(t *testing.T) {
	svc := &fakeEditService{returnedProfile: sampleProfileOutput}
	r := newEditTestRouter(svc)
	w := doPatchProfile(r, "1", `{"name":"Otro Nombre","depth_level":"basico"}`)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Data struct {
			Profile service.ProfileOutput `json:"profile"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if resp.Data.Profile.UserID != sampleProfileOutput.UserID {
		t.Fatalf("expected user_id %d, got %d", sampleProfileOutput.UserID, resp.Data.Profile.UserID)
	}
}

func TestPatchProfile_SinCampos(t *testing.T) {
	r := newEditTestRouter(&fakeEditService{})
	w := doPatchProfile(r, "1", `{}`)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestPatchProfile_UserIDInvalido(t *testing.T) {
	r := newEditTestRouter(&fakeEditService{})
	w := doPatchProfile(r, "abc", `{"name":"Nombre"}`)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestPatchProfile_ErrorServicio(t *testing.T) {
	r := newEditTestRouter(&fakeEditService{updateErr: true})
	w := doPatchProfile(r, "1", `{"name":"Nombre"}`)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d — body: %s", w.Code, w.Body.String())
	}
}

// ─── POST /users/:user_id/interests ──────────────────────────────────────────

func doAddInterest(r *gin.Engine, userID string, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/"+userID+"/interests", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestAddInterest_Success(t *testing.T) {
	svc := &fakeEditService{returnedProfile: sampleProfileOutput}
	r := newEditTestRouter(svc)
	w := doAddInterest(r, "1", `{"category_id":2}`)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
	if svc.capturedUserID != 1 {
		t.Fatalf("expected user_id 1, got %d", svc.capturedUserID)
	}
	if svc.capturedItemID != 2 {
		t.Fatalf("expected category_id 2, got %d", svc.capturedItemID)
	}

	var resp struct {
		Data struct {
			Profile service.ProfileOutput `json:"profile"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if resp.Data.Profile.UserID != sampleProfileOutput.UserID {
		t.Fatalf("expected user_id %d, got %d", sampleProfileOutput.UserID, resp.Data.Profile.UserID)
	}
}

func TestAddInterest_SinCategoryID(t *testing.T) {
	r := newEditTestRouter(&fakeEditService{})
	w := doAddInterest(r, "1", `{}`)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestAddInterest_UserIDInvalido(t *testing.T) {
	r := newEditTestRouter(&fakeEditService{})
	w := doAddInterest(r, "xyz", `{"category_id":2}`)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestAddInterest_ErrorServicio(t *testing.T) {
	r := newEditTestRouter(&fakeEditService{addInterestErr: true})
	w := doAddInterest(r, "1", `{"category_id":2}`)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d — body: %s", w.Code, w.Body.String())
	}
}

// ─── DELETE /users/:user_id/interests/:category_id ───────────────────────────

func doRemoveInterest(r *gin.Engine, userID string, categoryID string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/users/"+userID+"/interests/"+categoryID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestRemoveInterest_Success(t *testing.T) {
	svc := &fakeEditService{returnedProfile: sampleProfileOutput}
	r := newEditTestRouter(svc)
	w := doRemoveInterest(r, "1", "2")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
	if svc.capturedUserID != 1 {
		t.Fatalf("expected user_id 1, got %d", svc.capturedUserID)
	}
	if svc.capturedItemID != 2 {
		t.Fatalf("expected category_id 2, got %d", svc.capturedItemID)
	}
}

func TestRemoveInterest_UserIDInvalido(t *testing.T) {
	r := newEditTestRouter(&fakeEditService{})
	w := doRemoveInterest(r, "abc", "2")

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestRemoveInterest_CategoryIDInvalido(t *testing.T) {
	r := newEditTestRouter(&fakeEditService{})
	w := doRemoveInterest(r, "1", "abc")

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestRemoveInterest_ErrorServicio(t *testing.T) {
	r := newEditTestRouter(&fakeEditService{remInterestErr: true})
	w := doRemoveInterest(r, "1", "2")

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d — body: %s", w.Code, w.Body.String())
	}
}

// ─── POST /users/:user_id/focus-types ────────────────────────────────────────

func doAddFocusType(r *gin.Engine, userID string, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/"+userID+"/focus-types", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestAddFocusType_Success(t *testing.T) {
	svc := &fakeEditService{returnedProfile: sampleProfileOutput}
	r := newEditTestRouter(svc)
	w := doAddFocusType(r, "1", `{"focus_type_id":3}`)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
	if svc.capturedUserID != 1 {
		t.Fatalf("expected user_id 1, got %d", svc.capturedUserID)
	}
	if svc.capturedItemID != 3 {
		t.Fatalf("expected focus_type_id 3, got %d", svc.capturedItemID)
	}

	var resp struct {
		Data struct {
			Profile service.ProfileOutput `json:"profile"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if resp.Data.Profile.UserID != sampleProfileOutput.UserID {
		t.Fatalf("expected user_id %d, got %d", sampleProfileOutput.UserID, resp.Data.Profile.UserID)
	}
}

func TestAddFocusType_SinFocusTypeID(t *testing.T) {
	r := newEditTestRouter(&fakeEditService{})
	w := doAddFocusType(r, "1", `{}`)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestAddFocusType_UserIDInvalido(t *testing.T) {
	r := newEditTestRouter(&fakeEditService{})
	w := doAddFocusType(r, "xyz", `{"focus_type_id":1}`)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestAddFocusType_ErrorServicio(t *testing.T) {
	r := newEditTestRouter(&fakeEditService{addFocusErr: true})
	w := doAddFocusType(r, "1", `{"focus_type_id":1}`)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d — body: %s", w.Code, w.Body.String())
	}
}

// ─── DELETE /users/:user_id/focus-types/:focus_type_id ───────────────────────

func doRemoveFocusType(r *gin.Engine, userID string, focusTypeID string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/users/"+userID+"/focus-types/"+focusTypeID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestRemoveFocusType_Success(t *testing.T) {
	svc := &fakeEditService{returnedProfile: sampleProfileOutput}
	r := newEditTestRouter(svc)
	w := doRemoveFocusType(r, "1", "3")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
	if svc.capturedUserID != 1 {
		t.Fatalf("expected user_id 1, got %d", svc.capturedUserID)
	}
	if svc.capturedItemID != 3 {
		t.Fatalf("expected focus_type_id 3, got %d", svc.capturedItemID)
	}
}

func TestRemoveFocusType_UserIDInvalido(t *testing.T) {
	r := newEditTestRouter(&fakeEditService{})
	w := doRemoveFocusType(r, "abc", "3")

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestRemoveFocusType_FocusTypeIDInvalido(t *testing.T) {
	r := newEditTestRouter(&fakeEditService{})
	w := doRemoveFocusType(r, "1", "abc")

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestRemoveFocusType_ErrorServicio(t *testing.T) {
	r := newEditTestRouter(&fakeEditService{remFocusErr: true})
	w := doRemoveFocusType(r, "1", "3")

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d — body: %s", w.Code, w.Body.String())
	}
}

// ─── GET /users/:user_id ──────────────────────────────────────────────────────

func doGetProfile(r *gin.Engine, userID string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/"+userID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestGetProfile_Success(t *testing.T) {
	svc := &fakeEditService{returnedProfile: sampleProfileOutput}
	r := newEditTestRouter(svc)
	w := doGetProfile(r, "1")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
	if svc.capturedUserID != 1 {
		t.Fatalf("expected user_id 1, got %d", svc.capturedUserID)
	}

	var resp struct {
		Data struct {
			Profile service.ProfileOutput `json:"profile"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if resp.Data.Profile.UserID != sampleProfileOutput.UserID {
		t.Fatalf("expected user_id %d, got %d", sampleProfileOutput.UserID, resp.Data.Profile.UserID)
	}
	if resp.Data.Profile.Name != sampleProfileOutput.Name {
		t.Fatalf("expected name %q, got %q", sampleProfileOutput.Name, resp.Data.Profile.Name)
	}
}

func TestGetProfile_UserIDInvalido(t *testing.T) {
	r := newEditTestRouter(&fakeEditService{})
	w := doGetProfile(r, "abc")

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestGetProfile_ErrorServicio(t *testing.T) {
	r := newEditTestRouter(&fakeEditService{getProfileErr: true})
	w := doGetProfile(r, "1")

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d — body: %s", w.Code, w.Body.String())
	}
}
