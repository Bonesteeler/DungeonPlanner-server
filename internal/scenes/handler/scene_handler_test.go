package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"dungeonplannerserver/dto"
	authmodel "dungeonplannerserver/internal/auth/model"
	"dungeonplannerserver/internal/scenes/model"
)

// --- Mock ---

type mockSceneService struct {
	listScenesFn                  func(offset int) ([]model.Scene, int)
	listPendingScenesFn           func(offset int) ([]model.Scene, int, error)
	getSceneByIDFn                func(id uuid.UUID) *model.Scene
	addSceneFn                    func(scene model.Scene) error
	updateSceneModerationStatusFn func(id uuid.UUID, status dto.ModerationStatus) (bool, error)
	getSceneStatsFn               func() int
}

func (m *mockSceneService) ListScenes(offset int) ([]model.Scene, int) { return m.listScenesFn(offset) }
func (m *mockSceneService) ListPendingScenes(offset int) ([]model.Scene, int, error) {
	return m.listPendingScenesFn(offset)
}
func (m *mockSceneService) GetSceneByID(id uuid.UUID) *model.Scene { return m.getSceneByIDFn(id) }
func (m *mockSceneService) AddScene(scene model.Scene) error       { return m.addSceneFn(scene) }
func (m *mockSceneService) UpdateSceneModerationStatus(id uuid.UUID, status dto.ModerationStatus) (bool, error) {
	return m.updateSceneModerationStatusFn(id, status)
}
func (m *mockSceneService) GetSceneStats() int { return m.getSceneStatsFn() }

// --- Helpers ---

func newEchoContext(method, path string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

// --- ListScenes ---

func TestListScenes_ReturnsScenes(t *testing.T) {
	id := uuid.New()
	h := NewSceneHandler(&mockSceneService{
		listScenesFn: func(offset int) ([]model.Scene, int) {
			return []model.Scene{{ID: id.String(), Name: "Scene B", Author: "Bob"}}, 1
		},
	})

	c, rec := newEchoContext(http.MethodGet, "/scenes/list/0")
	if err := h.ListScenes(c, 0); err != nil {
		t.Fatalf("ListScenes() returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var got dto.SceneListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if len(got.Scenes) != 1 || got.Scenes[0].Name != "Scene B" {
		t.Errorf("unexpected response body: %v", got)
	}
	if got.TotalCount != 1 {
		t.Errorf("TotalCount = %d, want 1", got.TotalCount)
	}
	if got.PageSize != pageSize {
		t.Errorf("PageSize = %d, want %d", got.PageSize, pageSize)
	}
}

// --- GetSceneByID ---

func TestGetSceneByID_ValidUUID_ReturnsScene(t *testing.T) {
	id := uuid.New()
	h := NewSceneHandler(&mockSceneService{
		getSceneByIDFn: func(got uuid.UUID) *model.Scene {
			if got != id {
				t.Errorf("GetSceneByID called with %v, want %v", got, id)
			}
			return &model.Scene{ID: id.String(), Name: "Cave", Author: "Carol"}
		},
	})

	c, rec := newEchoContext(http.MethodGet, "/scenes/"+id.String())
	if err := h.GetSceneByID(c, id.String()); err != nil {
		t.Fatalf("GetSceneByID() returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var got dto.SceneResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if got.Name != "Cave" {
		t.Errorf("response Name = %s, want Cave", got.Name)
	}
}

func TestGetSceneByID_InvalidUUID_ReturnsBadRequest(t *testing.T) {
	h := NewSceneHandler(&mockSceneService{})

	c, rec := newEchoContext(http.MethodGet, "/scenes/not-a-uuid")
	if err := h.GetSceneByID(c, "not-a-uuid"); err != nil {
		t.Fatalf("GetSceneByID() returned error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	var got struct{ Error string }
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if got.Error != "Invalid UUID format" {
		t.Errorf("error message = %q, want %q", got.Error, "Invalid UUID format")
	}
}

// --- AddScene ---

func TestAddScene_ReturnsSuccess(t *testing.T) {
	called := false
	h := NewSceneHandler(&mockSceneService{
		getSceneByIDFn: func(id uuid.UUID) *model.Scene { return nil },
		addSceneFn:     func(scene model.Scene) error { called = true; return nil },
	})

	c, rec := newEchoContext(http.MethodPost, "/scenes/add")
	c.Set("claims", &authmodel.Claims{UserID: uuid.New().String()})
	if err := h.AddScene(c, dto.AddSceneRequest{ID: uuid.New().String(), Name: "New Scene"}); err != nil {
		t.Fatalf("AddScene() returned error: %v", err)
	}

	if !called {
		t.Error("expected AddScene service method to be called")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var got struct{ Status string }
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if got.Status != "Scene added successfully" {
		t.Errorf("status message = %q, want %q", got.Status, "Scene added successfully")
	}
}

func TestAddScene_ForwardsRequestToService(t *testing.T) {
	var captured model.Scene
	h := NewSceneHandler(&mockSceneService{
		getSceneByIDFn: func(id uuid.UUID) *model.Scene { return nil },
		addSceneFn:     func(scene model.Scene) error { captured = scene; return nil },
	})

	sceneID := uuid.New()
	authorID := uuid.New()
	req := dto.AddSceneRequest{
		ID:     sceneID.String(),
		Name:   "Fortress",
		Layers: []dto.LayerRequest{{Tiles: []dto.TileRequest{{TileID: "t1", Rotation: 90, XPos: 1, YPos: 2}}}},
	}
	c, _ := newEchoContext(http.MethodPost, "/scenes/add")
	c.Set("claims", &authmodel.Claims{UserID: authorID.String()})
	h.AddScene(c, req)

	if captured.ID != sceneID.String() {
		t.Errorf("captured scene ID = %s, want %s", captured.ID, sceneID.String())
	}
	if captured.Name != "Fortress" || captured.AuthorID != authorID {
		t.Errorf("captured scene = %+v, want Name=Fortress AuthorID=%v", captured, authorID)
	}
	if len(captured.Layers) != 1 || len(captured.Layers[0].Tiles) != 1 || captured.Layers[0].Tiles[0].TileID != "t1" {
		t.Errorf("captured layers = %+v, want 1 layer with 1 tile TileID=t1", captured.Layers)
	}
}

func TestAddScene_MissingClaims_ReturnsUnauthorized(t *testing.T) {
	called := false
	h := NewSceneHandler(&mockSceneService{
		addSceneFn: func(scene model.Scene) error { called = true; return nil },
	})

	c, rec := newEchoContext(http.MethodPost, "/scenes/add")
	if err := h.AddScene(c, dto.AddSceneRequest{Name: "New Scene"}); err != nil {
		t.Fatalf("AddScene() returned error: %v", err)
	}

	if called {
		t.Error("expected AddScene service method not to be called")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAddScene_InvalidUserIDInClaims_ReturnsBadRequest(t *testing.T) {
	called := false
	h := NewSceneHandler(&mockSceneService{
		addSceneFn: func(scene model.Scene) error { called = true; return nil },
	})

	c, rec := newEchoContext(http.MethodPost, "/scenes/add")
	c.Set("claims", &authmodel.Claims{UserID: "not-a-uuid"})
	if err := h.AddScene(c, dto.AddSceneRequest{Name: "New Scene"}); err != nil {
		t.Fatalf("AddScene() returned error: %v", err)
	}

	if called {
		t.Error("expected AddScene service method not to be called")
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestAddScene_InvalidSceneID_ReturnsBadRequest(t *testing.T) {
	called := false
	h := NewSceneHandler(&mockSceneService{
		getSceneByIDFn: func(id uuid.UUID) *model.Scene { return nil },
		addSceneFn:     func(scene model.Scene) error { called = true; return nil },
	})

	c, rec := newEchoContext(http.MethodPost, "/scenes/add")
	c.Set("claims", &authmodel.Claims{UserID: uuid.New().String()})
	if err := h.AddScene(c, dto.AddSceneRequest{ID: "not-a-uuid", Name: "New Scene"}); err != nil {
		t.Fatalf("AddScene() returned error: %v", err)
	}

	if called {
		t.Error("expected AddScene service method not to be called")
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	var got struct{ Error string }
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if got.Error != "Invalid UUID format" {
		t.Errorf("error message = %q, want %q", got.Error, "Invalid UUID format")
	}
}

func TestAddScene_DuplicateID_ReturnsConflict(t *testing.T) {
	called := false
	sceneID := uuid.New()
	h := NewSceneHandler(&mockSceneService{
		getSceneByIDFn: func(id uuid.UUID) *model.Scene {
			return &model.Scene{ID: sceneID.String(), Name: "Existing Scene"}
		},
		addSceneFn: func(scene model.Scene) error { called = true; return nil },
	})

	c, rec := newEchoContext(http.MethodPost, "/scenes/add")
	c.Set("claims", &authmodel.Claims{UserID: uuid.New().String()})
	if err := h.AddScene(c, dto.AddSceneRequest{ID: sceneID.String(), Name: "New Scene"}); err != nil {
		t.Fatalf("AddScene() returned error: %v", err)
	}

	if called {
		t.Error("expected AddScene service method not to be called")
	}
	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
	var got struct{ Error string }
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if got.Error != "Scene with this ID already exists" {
		t.Errorf("error message = %q, want %q", got.Error, "Scene with this ID already exists")
	}
}

// --- ListPendingScenes ---

func TestListPendingScenes_ReturnsScenes(t *testing.T) {
	id := uuid.New()
	h := NewSceneHandler(&mockSceneService{
		listPendingScenesFn: func(offset int) ([]model.Scene, int, error) {
			if offset != 20 {
				t.Errorf("offset = %d, want 20", offset)
			}
			return []model.Scene{{ID: id.String(), Name: "Pending", Author: "Author"}}, 21, nil
		},
	})
	c, rec := newEchoContext(http.MethodGet, "/v1/scenes/moderation/pending/20")

	if err := h.ListPendingScenes(c, 20); err != nil {
		t.Fatalf("ListPendingScenes() returned error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var got dto.SceneListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if len(got.Scenes) != 1 || got.Scenes[0].ID != id.String() || got.TotalCount != 21 || got.PageSize != pageSize {
		t.Errorf("unexpected response: %+v", got)
	}
}

func TestListPendingScenes_ServiceError(t *testing.T) {
	h := NewSceneHandler(&mockSceneService{
		listPendingScenesFn: func(offset int) ([]model.Scene, int, error) {
			return nil, 0, errors.New("db error")
		},
	})
	c, rec := newEchoContext(http.MethodGet, "/v1/scenes/moderation/pending/0")

	if err := h.ListPendingScenes(c, 0); err != nil {
		t.Fatalf("ListPendingScenes() returned error: %v", err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

// --- UpdateSceneModerationStatus ---

func TestUpdateSceneModerationStatus_ApprovedAndRejected(t *testing.T) {
	sceneID := uuid.New()
	for _, status := range []dto.ModerationStatus{dto.ModerationStatusApproved, dto.ModerationStatusRejected} {
		called := false
		h := NewSceneHandler(&mockSceneService{
			updateSceneModerationStatusFn: func(id uuid.UUID, gotStatus dto.ModerationStatus) (bool, error) {
				called = true
				if id != sceneID || gotStatus != status {
					t.Errorf("got ID %v and status %v, want %v and %v", id, gotStatus, sceneID, status)
				}
				return true, nil
			},
		})
		c, rec := newEchoContext(http.MethodPatch, "/v1/scenes/moderation/"+sceneID.String())

		if err := h.UpdateSceneModerationStatus(c, sceneID.String(), dto.UpdateSceneModerationRequest{Status: status}); err != nil {
			t.Fatalf("UpdateSceneModerationStatus() returned error: %v", err)
		}
		if !called {
			t.Error("expected service to be called")
		}
		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}
	}
}

func TestUpdateSceneModerationStatus_InvalidUUID(t *testing.T) {
	h := NewSceneHandler(&mockSceneService{})
	c, rec := newEchoContext(http.MethodPatch, "/v1/scenes/moderation/not-a-uuid")

	if err := h.UpdateSceneModerationStatus(c, "not-a-uuid", dto.UpdateSceneModerationRequest{Status: dto.ModerationStatusApproved}); err != nil {
		t.Fatalf("UpdateSceneModerationStatus() returned error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestUpdateSceneModerationStatus_InvalidStatus(t *testing.T) {
	h := NewSceneHandler(&mockSceneService{})
	for _, status := range []dto.ModerationStatus{dto.ModerationStatusPending, dto.ModerationStatus(99)} {
		c, rec := newEchoContext(http.MethodPatch, "/v1/scenes/moderation/"+uuid.NewString())
		if err := h.UpdateSceneModerationStatus(c, uuid.NewString(), dto.UpdateSceneModerationRequest{Status: status}); err != nil {
			t.Fatalf("UpdateSceneModerationStatus() returned error: %v", err)
		}
		if rec.Code != http.StatusBadRequest {
			t.Errorf("moderation status %v returned HTTP %d, want %d", status, rec.Code, http.StatusBadRequest)
		}
	}
}

func TestUpdateSceneModerationStatus_NoPendingSceneReturnsConflict(t *testing.T) {
	h := NewSceneHandler(&mockSceneService{
		updateSceneModerationStatusFn: func(id uuid.UUID, status dto.ModerationStatus) (bool, error) {
			return false, nil
		},
	})
	c, rec := newEchoContext(http.MethodPatch, "/v1/scenes/moderation/"+uuid.NewString())

	if err := h.UpdateSceneModerationStatus(c, uuid.NewString(), dto.UpdateSceneModerationRequest{Status: dto.ModerationStatusApproved}); err != nil {
		t.Fatalf("UpdateSceneModerationStatus() returned error: %v", err)
	}
	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
}

func TestUpdateSceneModerationStatus_ServiceError(t *testing.T) {
	h := NewSceneHandler(&mockSceneService{
		updateSceneModerationStatusFn: func(id uuid.UUID, status dto.ModerationStatus) (bool, error) {
			return false, errors.New("db error")
		},
	})
	c, rec := newEchoContext(http.MethodPatch, "/v1/scenes/moderation/"+uuid.NewString())

	if err := h.UpdateSceneModerationStatus(c, uuid.NewString(), dto.UpdateSceneModerationRequest{Status: dto.ModerationStatusRejected}); err != nil {
		t.Fatalf("UpdateSceneModerationStatus() returned error: %v", err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}
