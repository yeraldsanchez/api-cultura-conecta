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

// fakeGroupService implementa transport.GroupService en memoria.
type fakeGroupService struct {
	err bool
}

func (f *fakeGroupService) CreateGroup(_ context.Context, input service.CreateGroupInput) (service.GroupOutput, error) {
	if f.err {
		return service.GroupOutput{}, errors.New("error de servicio")
	}
	return service.GroupOutput{
		ID:         1,
		Name:       input.Name,
		WorkID:     input.WorkID,
		CreatedBy:  input.CreatedBy,
		DepthLevel: input.DepthLevel,
	}, nil
}

func newGroupTestRouter(svc transport.GroupService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := transport.NewGroupHandler(svc)
	r := gin.New()
	r.POST("/api/v1/groups", h.CreateGroup)
	return r
}

func doCreateGroup(r *gin.Engine, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

const validGroupBody = `{
	"name":           "Grupo de Jazz",
	"work_id":        1,
	"created_by":     2,
	"depth_level":    "intermedio",
	"categories_ids": [1, 2]
}`

func TestCreateGroup_Success(t *testing.T) {
	r := newGroupTestRouter(&fakeGroupService{})
	w := doCreateGroup(r, validGroupBody)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d — body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Data struct {
			Group struct {
				ID   int32  `json:"id"`
				Name string `json:"name"`
			} `json:"group"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if resp.Data.Group.ID == 0 {
		t.Fatal("expected group id in response, got 0")
	}
	if resp.Data.Group.Name != "Grupo de Jazz" {
		t.Fatalf("expected name 'Grupo de Jazz', got %q", resp.Data.Group.Name)
	}
}

func TestCreateGroup_ServiceError(t *testing.T) {
	r := newGroupTestRouter(&fakeGroupService{err: true})
	w := doCreateGroup(r, validGroupBody)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestCreateGroup_CamposFaltantes(t *testing.T) {
	r := newGroupTestRouter(&fakeGroupService{})

	cases := []struct {
		name string
		body string
	}{
		{"sin name", `{"work_id":1,"created_by":2,"depth_level":"intermedio","categories_ids":[1]}`},
		{"sin work_id", `{"name":"Jazz","created_by":2,"depth_level":"intermedio","categories_ids":[1]}`},
		{"sin created_by", `{"name":"Jazz","work_id":1,"depth_level":"intermedio","categories_ids":[1]}`},
		{"sin depth_level", `{"name":"Jazz","work_id":1,"created_by":2,"categories_ids":[1]}`},
		{"sin categories_ids", `{"name":"Jazz","work_id":1,"created_by":2,"depth_level":"intermedio"}`},
		{"body vacío", `{}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := doCreateGroup(r, tc.body)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d — body: %s", w.Code, w.Body.String())
			}
		})
	}
}
