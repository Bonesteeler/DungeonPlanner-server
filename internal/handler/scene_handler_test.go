package handler

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/google/uuid"
    "github.com/labstack/echo/v4"

    "DungeonPlannerServer/internal/handler/dto"
)

// --- Mock ---

type mockSceneService struct {
    listScenesFn    func(offset int) []dto.SceneResponse
    getSceneByIDFn  func(id uuid.UUID) *dto.SceneResponse
    addSceneFn      func(scene dto.AddSceneRequest)
    getSceneStatsFn func() int
}

func (m *mockSceneService) ListScenes(offset int) []dto.SceneResponse   { return m.listScenesFn(offset) }
func (m *mockSceneService) GetSceneByID(id uuid.UUID) *dto.SceneResponse { return m.getSceneByIDFn(id) }
func (m *mockSceneService) AddScene(scene dto.AddSceneRequest)           { m.addSceneFn(scene) }
func (m *mockSceneService) GetSceneStats() int                           { return m.getSceneStatsFn() }

// --- Helpers ---

func newEchoContext(method, path string) (echo.Context, *httptest.ResponseRecorder) {
    e := echo.New()
    req := httptest.NewRequest(method, path, nil)
    rec := httptest.NewRecorder()
    return e.NewContext(req, rec), rec
}

// --- GetScenes ---

func TestGetScenes_ReturnsScenes(t *testing.T) {
    expected := []dto.SceneResponse{{ID: uuid.New().String(), Name: "Test Scene", Author: "Alice"}}
    h := NewSceneHandler(&mockSceneService{
        listScenesFn: func(offset int) []dto.SceneResponse { return expected },
    })

    c, rec := newEchoContext(http.MethodGet, "/scenes/")
    if err := h.GetScenes(c); err != nil {
        t.Fatalf("GetScenes() returned error: %v", err)
    }

    if rec.Code != http.StatusOK {
        t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
    }
    var got []dto.SceneResponse
    if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
        t.Fatalf("failed to unmarshal response: %v", err)
    }
    if len(got) != 1 || got[0].Name != "Test Scene" {
        t.Errorf("unexpected response body: %v", got)
    }
}

func TestGetScenes_CallsListScenesWithOffsetZero(t *testing.T) {
    capturedOffset := -1
    h := NewSceneHandler(&mockSceneService{
        listScenesFn: func(offset int) []dto.SceneResponse {
            capturedOffset = offset
            return nil
        },
    })

    c, _ := newEchoContext(http.MethodGet, "/scenes/")
    h.GetScenes(c)

    if capturedOffset != 0 {
        t.Errorf("ListScenes called with offset %d, want 0", capturedOffset)
    }
}

// --- ListScenes ---

func TestListScenes_ReturnsScenes(t *testing.T) {
    expected := []dto.SceneResponse{{ID: uuid.New().String(), Name: "Scene B", Author: "Bob"}}
    h := NewSceneHandler(&mockSceneService{
        listScenesFn: func(offset int) []dto.SceneResponse { return expected },
    })

    c, rec := newEchoContext(http.MethodGet, "/scenes/list/0")
    if err := h.ListScenes(c); err != nil {
        t.Fatalf("ListScenes() returned error: %v", err)
    }

    if rec.Code != http.StatusOK {
        t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
    }
    var got []dto.SceneResponse
    if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
        t.Fatalf("failed to unmarshal response: %v", err)
    }
    if len(got) != 1 || got[0].Name != "Scene B" {
        t.Errorf("unexpected response body: %v", got)
    }
}

// --- GetSceneByID ---

func TestGetSceneByID_ValidUUID_ReturnsScene(t *testing.T) {
    id := uuid.New()
    expected := &dto.SceneResponse{ID: id.String(), Name: "Cave", Author: "Carol"}
    h := NewSceneHandler(&mockSceneService{
        getSceneByIDFn: func(got uuid.UUID) *dto.SceneResponse {
            if got != id {
                t.Errorf("GetSceneByID called with %v, want %v", got, id)
            }
            return expected
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
        addSceneFn: func(scene dto.AddSceneRequest) { called = true },
    })

    c, rec := newEchoContext(http.MethodPost, "/scenes/add")
    if err := h.AddScene(c, dto.AddSceneRequest{Name: "New Scene", Author: "Dave"}); err != nil {
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
    var captured dto.AddSceneRequest
    h := NewSceneHandler(&mockSceneService{
        addSceneFn: func(scene dto.AddSceneRequest) { captured = scene },
    })

    req := dto.AddSceneRequest{
        Name:   "Fortress",
        Author: "Eve",
        Layers: []dto.LayerRequest{{Tiles: []dto.TileRequest{{TileID: "t1", Rotation: 90, XPos: 1, YPos: 2}}}},
    }
    c, _ := newEchoContext(http.MethodPost, "/scenes/add")
    h.AddScene(c, req)

    if captured.Name != "Fortress" || captured.Author != "Eve" {
        t.Errorf("captured request = %+v, want Name=Fortress Author=Eve", captured)
    }
    if len(captured.Layers) != 1 || len(captured.Layers[0].Tiles) != 1 || captured.Layers[0].Tiles[0].TileID != "t1" {
        t.Errorf("captured layers = %+v, want 1 layer with 1 tile TileID=t1", captured.Layers)
    }
}