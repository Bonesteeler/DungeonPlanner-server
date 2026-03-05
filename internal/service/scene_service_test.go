package service

import (
    "errors"
    "testing"

    "github.com/google/uuid"

    "DungeonPlannerServer/internal/db/tables"
    "DungeonPlannerServer/internal/handler/dto"
)

// --- Mock ---

type mockSceneRepo struct {
    listApprovedScenesFn func(offset int) ([]tables.Scene, error)
    getSceneByIDFn       func(id uuid.UUID) (*tables.Scene, error)
    addSceneFn           func(scene tables.Scene) error
}

func (m *mockSceneRepo) ListApprovedScenes(offset int) ([]tables.Scene, error) {
    return m.listApprovedScenesFn(offset)
}

func (m *mockSceneRepo) GetSceneByID(id uuid.UUID) (*tables.Scene, error) {
    return m.getSceneByIDFn(id)
}

func (m *mockSceneRepo) AddScene(scene tables.Scene) error {
    return m.addSceneFn(scene)
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
        listApprovedScenesFn: func(offset int) ([]tables.Scene, error) {
            return []tables.Scene{
                {ID: id1, Name: strPtr("Dungeon A"), Author: strPtr("Alice"), UniqueTileIDs: []string{"t1", "t2"}},
                {ID: id2, Name: strPtr("Dungeon B"), Author: strPtr("Bob"), UniqueTileIDs: []string{"t3"}},
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
        listApprovedScenesFn: func(offset int) ([]tables.Scene, error) {
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
        listApprovedScenesFn: func(offset int) ([]tables.Scene, error) {
            return []tables.Scene{}, nil
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
        getSceneByIDFn: func(id uuid.UUID) (*tables.Scene, error) {
            return &tables.Scene{
                ID:            sceneID,
                Name:          strPtr("Cave"),
                Author:        strPtr("Charlie"),
                UniqueTileIDs: []string{"t1", "t2"},
                Tiles: []tables.Tile{
                    {TileId: strPtr("t1"), Rotation: 90, XPos: 1, YPos: 2},
                    {TileId: strPtr("t2"), Rotation: 0, XPos: 3, YPos: 4},
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
    if len(result.Tiles) != 2 {
        t.Fatalf("expected 2 tiles, got %d", len(result.Tiles))
    }
    if result.Tiles[0].TileID != "t1" || result.Tiles[0].Rotation != 90 || result.Tiles[0].XPos != 1 || result.Tiles[0].YPos != 2 {
        t.Errorf("tile[0] mismatch: %+v", result.Tiles[0])
    }
    if result.Tiles[1].TileID != "t2" || result.Tiles[1].Rotation != 0 || result.Tiles[1].XPos != 3 || result.Tiles[1].YPos != 4 {
        t.Errorf("tile[1] mismatch: %+v", result.Tiles[1])
    }
}

func TestGetSceneByID_Success_NoTiles(t *testing.T) {
    sceneID := uuid.New()
    mock := &mockSceneRepo{
        getSceneByIDFn: func(id uuid.UUID) (*tables.Scene, error) {
            return &tables.Scene{
                ID:            sceneID,
                Name:          strPtr("Empty Room"),
                Author:        strPtr("Dave"),
                UniqueTileIDs: []string{},
                Tiles:         []tables.Tile{},
            }, nil
        },
    }
    sceneService := NewSceneService(mock)

    result := sceneService.GetSceneByID(sceneID)

    if result == nil {
        t.Fatal("expected non-nil result")
    }
    if len(result.Tiles) != 0 {
        t.Errorf("expected 0 tiles, got %d", len(result.Tiles))
    }
}

func TestGetSceneByID_RepoError_ReturnsNil(t *testing.T) {
    mock := &mockSceneRepo{
        getSceneByIDFn: func(id uuid.UUID) (*tables.Scene, error) {
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
        getSceneByIDFn: func(id uuid.UUID) (*tables.Scene, error) {
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
    var captured tables.Scene
    mock := &mockSceneRepo{
        addSceneFn: func(scene tables.Scene) error {
            captured = scene
            return nil
        },
    }
    sceneService := NewSceneService(mock)

    request := dto.AddSceneRequest{
        Name:   "Forest",
        Author: "Eve",
        Tiles: []dto.TileRequest{
            {TileID: "t1", Rotation: 0, XPos: 0, YPos: 0},
            {TileID: "t2", Rotation: 180, XPos: 1, YPos: 1},
            {TileID: "t1", Rotation: 90, XPos: 2, YPos: 2}, // duplicate tile ID
        },
    }

    sceneService.AddScene(request)

    if *captured.Name != "Forest" {
        t.Errorf("Name = %s, want Forest", *captured.Name)
    }
    if *captured.Author != "Eve" {
        t.Errorf("Author = %s, want Eve", *captured.Author)
    }
    if len(captured.UniqueTileIDs) != 2 {
        t.Errorf("UniqueTileIDs len = %d, want 2", len(captured.UniqueTileIDs))
    }
    if len(captured.Tiles) != 3 {
        t.Errorf("Tiles len = %d, want 3", len(captured.Tiles))
    }
}

func TestAddScene_RepoError_DoesNotPanic(t *testing.T) {
    mock := &mockSceneRepo{
        addSceneFn: func(scene tables.Scene) error {
            return errors.New("db write failed")
        },
    }
    sceneService := NewSceneService(mock)

    request := dto.AddSceneRequest{
        Name:   "Broken",
        Author: "Frank",
        Tiles:  []dto.TileRequest{{TileID: "t1", Rotation: 0, XPos: 0, YPos: 0}},
    }

    sceneService.AddScene(request) // should not panic
}

func TestAddScene_EmptyTiles(t *testing.T) {
    var captured tables.Scene
    mock := &mockSceneRepo{
        addSceneFn: func(scene tables.Scene) error {
            captured = scene
            return nil
        },
    }
    sceneService := NewSceneService(mock)

    sceneService.AddScene(dto.AddSceneRequest{Name: "Empty", Author: "Grace", Tiles: []dto.TileRequest{}})

    if len(captured.UniqueTileIDs) != 0 {
        t.Errorf("UniqueTileIDs len = %d, want 0", len(captured.UniqueTileIDs))
    }
    if len(captured.Tiles) != 0 {
        t.Errorf("Tiles len = %d, want 0", len(captured.Tiles))
    }
}

// --- _convertTileRequestsToTiles ---

func TestConvertTileRequestsToTiles(t *testing.T) {
    requests := []dto.TileRequest{
        {TileID: "t1", Rotation: 90, XPos: 5, YPos: 10},
        {TileID: "t2", Rotation: 270, XPos: 3, YPos: 7},
    }

    tiles := _convertTileRequestsToTiles(requests)

    if len(tiles) != 2 {
        t.Fatalf("expected 2 tiles, got %d", len(tiles))
    }
    if *tiles[0].TileId != "t1" || tiles[0].Rotation != 90 || tiles[0].XPos != 5 || tiles[0].YPos != 10 {
        t.Errorf("tile[0] mismatch: %+v", tiles[0])
    }
    if *tiles[1].TileId != "t2" || tiles[1].Rotation != 270 || tiles[1].XPos != 3 || tiles[1].YPos != 7 {
        t.Errorf("tile[1] mismatch: %+v", tiles[1])
    }
}

func TestConvertTileRequestsToTiles_Empty(t *testing.T) {
    tiles := _convertTileRequestsToTiles([]dto.TileRequest{})
    if len(tiles) != 0 {
        t.Errorf("expected 0 tiles, got %d", len(tiles))
    }
}