package transport_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"api-cultura-conecta/internal/apperrors"
	"api-cultura-conecta/internal/service"
	"api-cultura-conecta/internal/transport"

	"github.com/gin-gonic/gin"
)

// fakeGroupService implementa transport.GroupService en memoria.
type fakeGroupService struct {
	err                bool
	listResult         service.ListGroupsOutput
	listErr            bool
	capturedInput      service.ListGroupsInput
	joinErr            error
	createPostErr      error
	createPostResult   service.PostOutput
	userGroupsResult   []service.UserGroupOutput
	userGroupsErr      error
	groupMembersResult []service.GroupMemberOutput
	groupMembersErr    error
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

func (f *fakeGroupService) ListGroups(_ context.Context, input service.ListGroupsInput) (service.ListGroupsOutput, error) {
	f.capturedInput = input
	if f.listErr {
		return service.ListGroupsOutput{}, errors.New("error de servicio")
	}
	return f.listResult, nil
}

func (f *fakeGroupService) JoinGroup(_ context.Context, _ int32, _ int32) error {
	return f.joinErr
}

func (f *fakeGroupService) CreatePost(_ context.Context, _ service.CreatePostInput) (service.PostOutput, error) {
	return f.createPostResult, f.createPostErr
}

func (f *fakeGroupService) GetSuggestedGroups(_ context.Context, _ service.SuggestGroupsInput) (service.ListGroupsOutput, error) {
	if f.listErr {
		return service.ListGroupsOutput{}, errors.New("error de servicio")
	}
	return f.listResult, nil
}

func (f *fakeGroupService) GetGroupsByMember(_ context.Context, _ int32) ([]service.UserGroupOutput, error) {
	return f.userGroupsResult, f.userGroupsErr
}

func (f *fakeGroupService) GetGroupMembers(_ context.Context, _ int32) ([]service.GroupMemberOutput, error) {
	return f.groupMembersResult, f.groupMembersErr
}

func (f *fakeGroupService) ListGroupPosts(_ context.Context, _ service.ListGroupPostsInput) (service.ListGroupPostsOutput, error) {
	return service.ListGroupPostsOutput{Posts: []service.PostWithAuthorOutput{}}, nil
}

func newGroupTestRouter(svc transport.GroupService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := transport.NewGroupHandler(svc)
	r := gin.New()
	r.GET("/api/v1/groups", h.ListGroups)
	r.POST("/api/v1/groups", h.CreateGroup)
	return r
}

func newGroupTestRouterWithAuth(svc transport.GroupService, userID int32) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := transport.NewGroupHandler(svc)
	r := gin.New()
	r.POST("/api/v1/groups/:group_id/members", func(c *gin.Context) {
		c.Set(transport.UserIDKey, userID)
		c.Next()
	}, h.JoinGroup)
	return r
}

func doJoinGroup(r *gin.Engine, groupID string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups/"+groupID+"/members", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func doListGroups(r *gin.Engine, query string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/groups"+query, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

var sampleGroups = []service.GroupOutput{
	{
		ID:         1,
		Name:       "Grupo de Jazz",
		WorkID:     1,
		WorkTitle:  "Kind of Blue",
		CreatedBy:  2,
		DepthLevel: "intermedio",
		Interests:  []service.InterestOutput{{ID: 1, Name: "Música"}},
	},
	{
		ID:         2,
		Name:       "Club de Lectura",
		WorkID:     2,
		WorkTitle:  "Cien años de soledad",
		CreatedBy:  3,
		DepthLevel: "avanzado",
		Interests:  []service.InterestOutput{{ID: 2, Name: "Literatura"}},
	},
}

func TestListGroups_SinFiltros(t *testing.T) {
	svc := &fakeGroupService{
		listResult: service.ListGroupsOutput{Groups: sampleGroups, Total: 2},
	}
	r := newGroupTestRouter(svc)
	w := doListGroups(r, "")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Data struct {
			Items      []service.GroupOutput `json:"items"`
			TotalCount int64                 `json:"total_count"`
			Page       int32                 `json:"page"`
			Limit      int32                 `json:"limit"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(resp.Data.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(resp.Data.Items))
	}
	if resp.Data.TotalCount != 2 {
		t.Fatalf("expected total_count 2, got %d", resp.Data.TotalCount)
	}
	if resp.Data.Page != 1 {
		t.Fatalf("expected page 1, got %d", resp.Data.Page)
	}
	if svc.capturedInput.WorkID != nil || svc.capturedInput.Name != nil || svc.capturedInput.DepthLevel != nil {
		t.Fatal("expected no filters to be set")
	}
}

func TestListGroups_FiltrosCombinados(t *testing.T) {
	filtered := sampleGroups[:1]
	svc := &fakeGroupService{
		listResult: service.ListGroupsOutput{Groups: filtered, Total: 1},
	}
	r := newGroupTestRouter(svc)
	w := doListGroups(r, "?name=Jazz&depth_level=intermedio&categories_ids=1&work_id=1&page=1&limit=5")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Data struct {
			Items      []service.GroupOutput `json:"items"`
			TotalCount int64                 `json:"total_count"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(resp.Data.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(resp.Data.Items))
	}
	if resp.Data.Items[0].Name != "Grupo de Jazz" {
		t.Fatalf("expected 'Grupo de Jazz', got %q", resp.Data.Items[0].Name)
	}

	in := svc.capturedInput
	if in.Name == nil || *in.Name != "Jazz" {
		t.Fatalf("expected name filter 'Jazz', got %v", in.Name)
	}
	if in.DepthLevel == nil || *in.DepthLevel != "intermedio" {
		t.Fatalf("expected depth_level 'intermedio', got %v", in.DepthLevel)
	}
	if in.WorkID == nil || *in.WorkID != 1 {
		t.Fatalf("expected work_id 1, got %v", in.WorkID)
	}
	if len(in.FocusTypeIDs) != 1 || in.FocusTypeIDs[0] != 1 {
		t.Fatalf("expected focus_type_ids [1], got %v", in.FocusTypeIDs)
	}
}

func TestListGroups_SinResultados(t *testing.T) {
	svc := &fakeGroupService{
		listResult: service.ListGroupsOutput{Groups: []service.GroupOutput{}, Total: 0},
	}
	r := newGroupTestRouter(svc)
	w := doListGroups(r, "?name=NoExiste")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Data struct {
			Items      []service.GroupOutput `json:"items"`
			TotalCount int64                 `json:"total_count"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(resp.Data.Items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(resp.Data.Items))
	}
	if resp.Data.TotalCount != 0 {
		t.Fatalf("expected total_count 0, got %d", resp.Data.TotalCount)
	}
}

func TestListGroups_ErrorServicio(t *testing.T) {
	svc := &fakeGroupService{listErr: true}
	r := newGroupTestRouter(svc)
	w := doListGroups(r, "")

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d — body: %s", w.Code, w.Body.String())
	}
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

func TestJoinGroup_Success(t *testing.T) {
	r := newGroupTestRouterWithAuth(&fakeGroupService{}, 5)
	w := doJoinGroup(r, "1")

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestJoinGroup_GrupoInexistente(t *testing.T) {
	svc := &fakeGroupService{joinErr: apperrors.ErrGroupNotFound}
	r := newGroupTestRouterWithAuth(svc, 5)
	w := doJoinGroup(r, "9999")

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestJoinGroup_YaEsMiembro(t *testing.T) {
	svc := &fakeGroupService{joinErr: apperrors.ErrAlreadyMember}
	r := newGroupTestRouterWithAuth(svc, 5)
	w := doJoinGroup(r, "1")

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestJoinGroup_ErrorServicio(t *testing.T) {
	svc := &fakeGroupService{joinErr: errors.New("error inesperado")}
	r := newGroupTestRouterWithAuth(svc, 5)
	w := doJoinGroup(r, "1")

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestJoinGroup_GroupIDInvalido(t *testing.T) {
	r := newGroupTestRouterWithAuth(&fakeGroupService{}, 5)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups/abc/members", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d — body: %s", w.Code, w.Body.String())
	}
}

func newPostTestRouter(svc transport.GroupService, userID int32) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := transport.NewGroupHandler(svc)
	r := gin.New()
	r.POST("/api/v1/groups/:group_id/posts", func(c *gin.Context) {
		c.Set(transport.UserIDKey, userID)
		c.Next()
	}, h.CreatePost)
	return r
}

func doCreatePost(r *gin.Engine, groupID string, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups/"+groupID+"/posts", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestCreatePost_Success(t *testing.T) {
	sp := "Capítulo 5"
	svc := &fakeGroupService{
		createPostResult: service.PostOutput{
			ID:              1,
			GroupID:         3,
			UserID:          7,
			Content:         "Mi opinión",
			HasSpoiler:      true,
			SpoilerProgress: &sp,
		},
	}
	r := newPostTestRouter(svc, 7)
	w := doCreatePost(r, "3", `{"content":"Mi opinión","has_spoiler":true,"spoiler_progress":"Capítulo 5"}`)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d — body: %s", w.Code, w.Body.String())
	}
	var resp struct {
		Data struct {
			Post service.PostOutput `json:"post"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if resp.Data.Post.ID != 1 {
		t.Fatalf("expected post id 1, got %d", resp.Data.Post.ID)
	}
}

func TestCreatePost_NotMember(t *testing.T) {
	svc := &fakeGroupService{createPostErr: apperrors.ErrNotGroupMember}
	r := newPostTestRouter(svc, 99)
	w := doCreatePost(r, "3", `{"content":"Hola"}`)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestCreatePost_InvalidBody(t *testing.T) {
	r := newPostTestRouter(&fakeGroupService{}, 7)
	w := doCreatePost(r, "3", `{}`)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestCreatePost_GroupNotFound(t *testing.T) {
	svc := &fakeGroupService{createPostErr: apperrors.ErrGroupNotFound}
	r := newPostTestRouter(svc, 7)
	w := doCreatePost(r, "9999", `{"content":"Hola"}`)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d — body: %s", w.Code, w.Body.String())
	}
}

func newSuggestionsTestRouter(svc transport.GroupService, userID int32) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := transport.NewGroupHandler(svc)
	r := gin.New()
	r.GET("/api/v1/groups/suggestions", func(c *gin.Context) {
		c.Set(transport.UserIDKey, userID)
		c.Next()
	}, h.GetSuggestedGroups)
	return r
}

func doGetSuggestions(r *gin.Engine, query string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/groups/suggestions"+query, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestGetSuggestedGroups_Success(t *testing.T) {
	svc := &fakeGroupService{
		listResult: service.ListGroupsOutput{Groups: sampleGroups, Total: 2},
	}
	r := newSuggestionsTestRouter(svc, 7)
	w := doGetSuggestions(r, "")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
	var resp struct {
		Data struct {
			Items      []service.GroupOutput `json:"items"`
			TotalCount int64                 `json:"total_count"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(resp.Data.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(resp.Data.Items))
	}
	if resp.Data.TotalCount != 2 {
		t.Fatalf("expected total_count 2, got %d", resp.Data.TotalCount)
	}
}

func TestGetSuggestedGroups_Empty(t *testing.T) {
	svc := &fakeGroupService{
		listResult: service.ListGroupsOutput{Groups: []service.GroupOutput{}, Total: 0},
	}
	r := newSuggestionsTestRouter(svc, 7)
	w := doGetSuggestions(r, "")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
	var resp struct {
		Data struct {
			Items      []service.GroupOutput `json:"items"`
			TotalCount int64                 `json:"total_count"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(resp.Data.Items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(resp.Data.Items))
	}
}

func TestGetSuggestedGroups_ServiceError(t *testing.T) {
	svc := &fakeGroupService{listErr: true}
	r := newSuggestionsTestRouter(svc, 7)
	w := doGetSuggestions(r, "")

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d — body: %s", w.Code, w.Body.String())
	}
}

func newUserGroupsTestRouter(svc transport.GroupService, userID int32) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := transport.NewGroupHandler(svc)
	r := gin.New()
	r.GET("/api/v1/users/:user_id/groups", func(c *gin.Context) {
		c.Set(transport.UserIDKey, userID)
		c.Next()
	}, h.GetGroupsByMember)
	return r
}

func newGroupMembersTestRouter(svc transport.GroupService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := transport.NewGroupHandler(svc)
	r := gin.New()
	r.GET("/api/v1/groups/:group_id/members", h.GetGroupMembers)
	return r
}

func TestGetGroupsByMember_Success(t *testing.T) {
	svc := &fakeGroupService{
		userGroupsResult: []service.UserGroupOutput{
			{ID: 1, Name: "Grupo de Jazz", Role: "member"},
			{ID: 2, Name: "Club de Lectura", Role: "admin"},
		},
	}
	r := newUserGroupsTestRouter(svc, 5)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/5/groups", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
	var resp struct {
		Data struct {
			Groups []service.UserGroupOutput `json:"groups"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(resp.Data.Groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(resp.Data.Groups))
	}
	if resp.Data.Groups[0].Role != "member" {
		t.Fatalf("expected role 'member', got %q", resp.Data.Groups[0].Role)
	}
}

func TestGetGroupsByMember_Empty(t *testing.T) {
	svc := &fakeGroupService{userGroupsResult: []service.UserGroupOutput{}}
	r := newUserGroupsTestRouter(svc, 5)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/5/groups", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestGetGroupsByMember_ServiceError(t *testing.T) {
	svc := &fakeGroupService{userGroupsErr: errors.New("db error")}
	r := newUserGroupsTestRouter(svc, 5)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/5/groups", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestGetGroupsByMember_InvalidUserID(t *testing.T) {
	r := newUserGroupsTestRouter(&fakeGroupService{}, 5)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/abc/groups", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestGetGroupMembers_Success(t *testing.T) {
	name1, name2 := "Ana García", "Luis Pérez"
	svc := &fakeGroupService{
		groupMembersResult: []service.GroupMemberOutput{
			{UserID: 1, Name: &name1, Role: "admin"},
			{UserID: 2, Name: &name2, Role: "member"},
		},
	}
	r := newGroupMembersTestRouter(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/groups/1/members", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
	var resp struct {
		Data struct {
			Members []service.GroupMemberOutput `json:"members"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(resp.Data.Members) != 2 {
		t.Fatalf("expected 2 members, got %d", len(resp.Data.Members))
	}
	if resp.Data.Members[0].Role != "admin" {
		t.Fatalf("expected first member role 'admin', got %q", resp.Data.Members[0].Role)
	}
}

func TestGetGroupMembers_Empty(t *testing.T) {
	svc := &fakeGroupService{groupMembersResult: []service.GroupMemberOutput{}}
	r := newGroupMembersTestRouter(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/groups/1/members", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestGetGroupMembers_ServiceError(t *testing.T) {
	svc := &fakeGroupService{groupMembersErr: errors.New("db error")}
	r := newGroupMembersTestRouter(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/groups/1/members", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestGetGroupMembers_InvalidGroupID(t *testing.T) {
	r := newGroupMembersTestRouter(&fakeGroupService{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/groups/abc/members", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d — body: %s", w.Code, w.Body.String())
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
