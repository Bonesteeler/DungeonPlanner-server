package service

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"DungeonPlannerServer/internal/handler/dto"
)

// --- Mock ---

type mockSceneRepo struct {
    listApprovedScenesFn func(offset int) ([]dto.SceneResponse, error)
    getSceneByIDFn       func(id uuid.UUID) (*dto.SceneResponse, error)
    addSceneFn           func(request dto.AddSceneRequest) error
}

func (m *mockSceneRepo) ListApprovedScenes(offset int) ([]dto.SceneResponse, error) {
    return m.listApprovedScenesFn(offset)
}

func (m *mockSceneRepo) GetSceneByID(id uuid.UUID) (*dto.SceneResponse, error) {
    return m.getSceneByIDFn(id)
}

func (m *mockSceneRepo) AddScene(request dto.AddSceneRequest) error {
    return m.addSceneFn(request)
}

// --- Helpers ---

func strPtr(s string) *string { return &s }

// --- GetSceneStats ---

func TestGetSceneStats_ReturnsZero(t *testing.T) {
    sceneService := NewSceneService(&mockSceneRepo{})
    if got := sceneService.GetSceneStats(); got != 0 {
        t.Errorf("GetSceneStats() = %d, want 0", got)
    }
}

// --- ListScenes ---

func TestListScenes_Success(t *testing.T) {
    id1 := uuid.New()
    id2 := uuid.New()
    mock := &mockSceneRepo{
        listApprovedScenesFn: func(offset int) ([]dto.SceneResponse, error) {
            return []dto.SceneResponse{
                {ID: id1.String(), Name: "Dungeon A", Author: "Alice", UniqueTileIDs: []string{"t1", "t2"}},
                {ID: id2.String(), Name: "Dungeon B", Author: "Bob", UniqueTileIDs: []string{"t3"}},
            }, nil
        },
    }
    sceneService := NewSceneService(mock)

    result := sceneService.ListScenes(5)

    if len(result) != 2 {
        t.Fatalf("expected 2 scenes, got %d", len(result))
    }
    if result[0].ID != id1.String() {
        t.Errorf("scene[0].ID = %s, want %s", result[0].ID, id1.String())
    }
    if result[0].Name != "Dungeon A" {
        t.Errorf("scene[0].Name = %s, want Dungeon A", result[0].Name)
    }
    if result[0].Author != "Alice" {
        t.Errorf("scene[0].Author = %s, want Alice", result[0].Author)
    }
    if len(result[0].UniqueTileIDs) != 2 {
        t.Errorf("scene[0].UniqueTileIDs len = %d, want 2", len(result[0].UniqueTileIDs))
    }
    if result[1].ID != id2.String() {
        t.Errorf("scene[1].ID = %s, want %s", result[1].ID, id2.String())
    }
}

func TestListScenes_RepoError_ReturnsEmpty(t *testing.T) {
    mock := &mockSceneRepo{
        listApprovedScenesFn: func(offset int) ([]dto.SceneResponse, error) {
            return nil, errors.New("db error")
        },
    }
    sceneService := NewSceneService(mock)

    result := sceneService.ListScenes(0)

    if len(result) != 0 {
        t.Errorf("expected empty slice on error, got %d items", len(result))
    }
}

func TestListScenes_EmptyResult(t *testing.T) {
    mock := &mockSceneRepo{
        listApprovedScenesFn: func(offset int) ([]dto.SceneResponse, error) {
            return []dto.SceneResponse{}, nil
        },
    }
    sceneService := NewSceneService(mock)

    result := sceneService.ListScenes(0)

    if len(result) != 0 {
        t.Errorf("expected empty/nil slice, got %d items", len(result))
    }
}

// --- GetSceneByID ---

func TestGetSceneByID_Success_WithTiles(t *testing.T) {
    sceneID := uuid.New()
    mock := &mockSceneRepo{
        getSceneByIDFn: func(id uuid.UUID) (*dto.SceneResponse, error) {
            return &dto.SceneResponse{
                ID:            sceneID.String(),
                Name:          "Cave",
                Author:        "Charlie",
                UniqueTileIDs: []string{"t1", "t2"},
                Layers: []dto.LayerResponse{
                    {Tiles: []dto.TileResponse{
                        {TileID: "t1", Rotation: 90, XPos: 1, YPos: 2},
                        {TileID: "t2", Rotation: 0, XPos: 3, YPos: 4},
                    }},
                },
            }, nil
        },
    }
    sceneService := NewSceneService(mock)

    result := sceneService.GetSceneByID(sceneID)

    if result == nil {
        t.Fatal("expected non-nil result")
    }
    if result.ID != sceneID.String() {
        t.Errorf("ID = %s, want %s", result.ID, sceneID.String())
    }
    if result.Name != "Cave" {
        t.Errorf("Name = %s, want Cave", result.Name)
    }
    if result.Author != "Charlie" {
        t.Errorf("Author = %s, want Charlie", result.Author)
    }
    if len(result.Layers) != 1 {
        t.Fatalf("expected 1 layer, got %d", len(result.Layers))
    }
    tiles := result.Layers[0].Tiles
    if len(tiles) != 2 {
        t.Fatalf("expected 2 tiles, got %d", len(tiles))
    }
    if tiles[0].TileID != "t1" || tiles[0].Rotation != 90 || tiles[0].XPos != 1 || tiles[0].YPos != 2 {
        t.Errorf("tile[0] mismatch: %+v", tiles[0])
    }
    if tiles[1].TileID != "t2" || tiles[1].Rotation != 0 || tiles[1].XPos != 3 || tiles[1].YPos != 4 {
        t.Errorf("tile[1] mismatch: %+v", tiles[1])
    }
}

func TestGetSceneByID_Success_NoTiles(t *testing.T) {
    sceneID := uuid.New()
    mock := &mockSceneRepo{
        getSceneByIDFn: func(id uuid.UUID) (*dto.SceneResponse, error) {
            return &dto.SceneResponse{
                ID:            sceneID.String(),
                Name:          "Empty Room",
                Author:        "Dave",
                UniqueTileIDs: []string{},
                Layers:        []dto.LayerResponse{},
            }, nil
        },
    }
    sceneService := NewSceneService(mock)

    result := sceneService.GetSceneByID(sceneID)

    if result == nil {
        t.Fatal("expected non-nil result")
    }
    if len(result.Layers) != 0 {
        t.Errorf("expected 0 layers, got %d", len(result.Layers))
    }
}

func TestGetSceneByID_RepoError_ReturnsNil(t *testing.T) {
    mock := &mockSceneRepo{
        getSceneByIDFn: func(id uuid.UUID) (*dto.SceneResponse, error) {
            return nil, errors.New("not found")
        },
    }
    sceneService := NewSceneService(mock)

    result := sceneService.GetSceneByID(uuid.New())

    if result != nil {
        t.Errorf("expected nil on error, got %+v", result)
    }
}

func TestGetSceneByID_NilScene_ReturnsNil(t *testing.T) {
    mock := &mockSceneRepo{
        getSceneByIDFn: func(id uuid.UUID) (*dto.SceneResponse, error) {
            return nil, nil
        },
    }
    sceneService := NewSceneService(mock)

    result := sceneService.GetSceneByID(uuid.New())

    if result != nil {
        t.Errorf("expected nil when scene not found, got %+v", result)
    }
}

// --- AddScene ---

func TestAddScene_Success(t *testing.T) {
    var captured dto.AddSceneRequest
    mock := &mockSceneRepo{
        addSceneFn: func(request dto.AddSceneRequest) error {
            captured = request
            return nil
        },
    }
    sceneService := NewSceneService(mock)

    request := dto.AddSceneRequest{
        Name:   "Forest",
        Author: "Eve",
        Layers: []dto.LayerRequest{
            {Tiles: []dto.TileRequest{
                {TileID: "t1", Rotation: 0, XPos: 0, YPos: 0},
                {TileID: "t2", Rotation: 180, XPos: 1, YPos: 1},
                {TileID: "t1", Rotation: 90, XPos: 2, YPos: 2},
            }},
        },
    }

    sceneService.AddScene(request)

    if captured.Name != "Forest" {
        t.Errorf("Name = %s, want Forest", captured.Name)
    }
    if captured.Author != "Eve" {
        t.Errorf("Author = %s, want Eve", captured.Author)
    }
}

func TestAddScene_RepoError_DoesNotPanic(t *testing.T) {
    mock := &mockSceneRepo{
        addSceneFn: func(request dto.AddSceneRequest) error {
            return errors.New("db write failed")
        },
    }
    sceneService := NewSceneService(mock)

    request := dto.AddSceneRequest{
        Name:   "Broken",
        Author: "Frank",
        Layers: []dto.LayerRequest{
            {Tiles: []dto.TileRequest{{TileID: "t1", Rotation: 0, XPos: 0, YPos: 0}}},
        },
    }

    sceneService.AddScene(request) // should not panic
}

func TestAddScene_EmptyLayers(t *testing.T) {
    var captured dto.AddSceneRequest
    mock := &mockSceneRepo{
        addSceneFn: func(request dto.AddSceneRequest) error {
            captured = request
            return nil
        },
    }
    sceneService := NewSceneService(mock)

    sceneService.AddScene(dto.AddSceneRequest{Name: "Empty", Author: "Grace", Layers: []dto.LayerRequest{}})

    if captured.Name != "Empty" {
        t.Errorf("Name = %s, want Empty", captured.Name)
    }
    if len(captured.Layers) != 0 {
        t.Errorf("Layers len = %d, want 0", len(captured.Layers))
    }
}

