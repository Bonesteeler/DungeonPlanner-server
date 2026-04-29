package repository

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/lib/pq"

	"DungeonPlannerServer/internal/db/tables"
	"DungeonPlannerServer/internal/model"
)

var sceneColumns = []string{"ID", "Name", "Author", "UniqueTileIDs", "ModerationStatus"}
var layerColumns = []string{"ID", "SceneId"}
var tileColumns = []string{"TileId", "Rotation", "XPos", "YPos", "LayerId"}

func getRepositoryWithMockDB(t *testing.T) (*SceneRepository, sqlmock.Sqlmock) {
    t.Helper()
    mockDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
    if err != nil {
        t.Fatalf("failed to create sqlmock: %v", err)
    }
    repo := NewSceneRepositoryWithDB(mockDB)
    t.Cleanup(func() {
        mockDB.Close()
    })
    return repo, mock
}

func strPtr(s string) *string { return &s }

func pqArrayValue(t *testing.T, ss []string) interface{} {
    t.Helper()
    v, err := pq.Array(ss).Value()
    if err != nil {
        t.Fatalf("pq.Array.Value() error: %v", err)
    }
    return v
}

// ---- GetApprovedSceneCount ----

func TestGetApprovedSceneCount_ReturnsCount(t *testing.T) {
    repo, mock := getRepositoryWithMockDB(t)

    mock.ExpectQuery(`SELECT COUNT(*) FROM public."Scenes" WHERE "ModerationStatus" = $1`).
        WithArgs(tables.ModerationStatusApproved).
        WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(7))

    count, err := repo.GetApprovedSceneCount()
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if count != 7 {
        t.Errorf("expected 7, got %d", count)
    }
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Error(err)
    }
}

func TestGetApprovedSceneCount_DBError(t *testing.T) {
    repo, mock := getRepositoryWithMockDB(t)

    mock.ExpectQuery(`SELECT COUNT(*) FROM public."Scenes" WHERE "ModerationStatus" = $1`).
        WithArgs(tables.ModerationStatusApproved).
        WillReturnError(errors.New("db error"))

    _, err := repo.GetApprovedSceneCount()
    if err == nil {
        t.Fatal("expected error, got nil")
    }
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Error(err)
    }
}

// ---- ListApprovedScenes ----

const listApprovedQuery = `SELECT "ID", "Name", "Author", "UniqueTileIDs", "ModerationStatus" FROM public."Scenes" WHERE "ModerationStatus" = $1 ORDER BY "ID" OFFSET $2 LIMIT $3`

func TestListApprovedScenes_ReturnsScenes(t *testing.T) {
    repo, mock := getRepositoryWithMockDB(t)

    id := uuid.New()
    rows := sqlmock.NewRows(sceneColumns).
        AddRow(id, strPtr("Test Scene"), strPtr("Author"), pqArrayValue(t, []string{"tile1", "tile2"}), tables.ModerationStatusApproved)

    mock.ExpectQuery(listApprovedQuery).
        WithArgs(tables.ModerationStatusApproved, 0, 20).
        WillReturnRows(rows)

    scenes, err := repo.ListApprovedScenes(0)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if len(scenes) != 1 {
        t.Fatalf("expected 1 scene, got %d", len(scenes))
    }
    if scenes[0].ID != id.String() {
        t.Errorf("expected ID %v, got %v", id.String(), scenes[0].ID)
    }
    if scenes[0].Name != "Test Scene" {
        t.Errorf("expected Name 'Test Scene', got '%v'", scenes[0].Name)
    }
    if scenes[0].Author != "Author" {
        t.Errorf("expected Author 'Author', got '%v'", scenes[0].Author)
    }
    if len(scenes[0].UniqueTileIDs) != 2 || scenes[0].UniqueTileIDs[0] != "tile1" || scenes[0].UniqueTileIDs[1] != "tile2" {
        t.Errorf("expected UniqueTileIDs ['tile1','tile2'], got %v", scenes[0].UniqueTileIDs)
    }
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Error(err)
    }
}

func TestListApprovedScenes_Empty(t *testing.T) {
    repo, mock := getRepositoryWithMockDB(t)

    mock.ExpectQuery(listApprovedQuery).
        WithArgs(tables.ModerationStatusApproved, 0, 20).
        WillReturnRows(sqlmock.NewRows(sceneColumns))

    scenes, err := repo.ListApprovedScenes(0)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if len(scenes) != 0 {
        t.Errorf("expected 0 scenes, got %d", len(scenes))
    }
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Error(err)
    }
}

func TestListApprovedScenes_AppliesOffsetPagination(t *testing.T) {
    repo, mock := getRepositoryWithMockDB(t)

    // Expect offset=40, limit=20 (third page)
    mock.ExpectQuery(listApprovedQuery).
        WithArgs(tables.ModerationStatusApproved, 40, 20).
        WillReturnRows(sqlmock.NewRows(sceneColumns))

    _, err := repo.ListApprovedScenes(40)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Error(err)
    }
}

func TestListApprovedScenes_DBError(t *testing.T) {
    repo, mock := getRepositoryWithMockDB(t)

    mock.ExpectQuery(listApprovedQuery).
        WillReturnError(errors.New("db error"))

    _, err := repo.ListApprovedScenes(0)
    if err == nil {
        t.Fatal("expected error, got nil")
    }
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Error(err)
    }
}

// ---- GetSceneByID ----

const getSceneByIDQuery = `SELECT "ID", "Name", "Author", "UniqueTileIDs", "ModerationStatus" FROM public."Scenes" WHERE "ID" = $1`
const getLayersBySceneIDQuery = `SELECT "ID", "SceneId" FROM public."Layers" WHERE "SceneId" = $1`
const getTilesByLayerIDQuery = `SELECT "TileId", "Rotation", "XPos", "YPos", "LayerId" FROM public."Tiles" WHERE "LayerId" = $1`

func TestGetSceneByID_Found(t *testing.T) {
    repo, mock := getRepositoryWithMockDB(t)

    sceneID := uuid.New()
    layerID := uuid.New()
    tileID := "tile-a"

    mock.ExpectQuery(getSceneByIDQuery).
        WithArgs(sceneID).
        WillReturnRows(sqlmock.NewRows(sceneColumns).
            AddRow(sceneID, strPtr("My Scene"), strPtr("Author"), pqArrayValue(t, []string{"tile-a"}), tables.ModerationStatusApproved))

    mock.ExpectQuery(getLayersBySceneIDQuery).
        WithArgs(sceneID).
        WillReturnRows(sqlmock.NewRows(layerColumns).
            AddRow(layerID, sceneID))

    mock.ExpectQuery(getTilesByLayerIDQuery).
        WithArgs(layerID).
        WillReturnRows(sqlmock.NewRows(tileColumns).
            AddRow(strPtr(tileID), 0, 1, 2, layerID))

    scene, err := repo.GetSceneByID(sceneID)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if scene == nil {
        t.Fatal("expected scene, got nil")
    }
    if scene.ID != sceneID.String() {
        t.Errorf("expected ID %v, got %v", sceneID.String(), scene.ID)
    }
    if len(scene.Layers) != 1 {
        t.Fatalf("expected 1 layer, got %d", len(scene.Layers))
    }
    if len(scene.Layers[0].Tiles) != 1 {
        t.Errorf("expected 1 tile, got %d", len(scene.Layers[0].Tiles))
    }
    if scene.Layers[0].Tiles[0].XPos != 1 || scene.Layers[0].Tiles[0].YPos != 2 {
        t.Errorf("expected tile position (1,2), got (%d,%d)", scene.Layers[0].Tiles[0].XPos, scene.Layers[0].Tiles[0].YPos)
    }
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Error(err)
    }
}

func TestGetSceneByID_NotFound(t *testing.T) {
    repo, mock := getRepositoryWithMockDB(t)
    id := uuid.New()

    mock.ExpectQuery(getSceneByIDQuery).
        WithArgs(id).
        WillReturnRows(sqlmock.NewRows(sceneColumns))

    scene, err := repo.GetSceneByID(id)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if scene != nil {
        t.Errorf("expected nil scene, got %+v", scene)
    }
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Error(err)
    }
}

func TestGetSceneByID_NoLayers(t *testing.T) {
    repo, mock := getRepositoryWithMockDB(t)
    id := uuid.New()
    mock.ExpectQuery(getSceneByIDQuery).
        WithArgs(id).
        WillReturnRows(sqlmock.NewRows(sceneColumns).
            AddRow(id, strPtr("Scene With No Layers"), strPtr("Author"), pqArrayValue(t, []string{}), tables.ModerationStatusApproved))
    mock.ExpectQuery(getLayersBySceneIDQuery).
        WithArgs(id).
        WillReturnRows(sqlmock.NewRows(layerColumns))

    scene, err := repo.GetSceneByID(id)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if scene == nil {
        t.Fatal("expected scene, got nil")
    }
    if len(scene.Layers) != 0 {
        t.Errorf("expected 0 layers, got %d", len(scene.Layers))
    }
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Error(err)
    }
}

func TestGetSceneByID_SceneQueryError(t *testing.T) {
    repo, mock := getRepositoryWithMockDB(t)
    id := uuid.New()

    mock.ExpectQuery(getSceneByIDQuery).
        WithArgs(id).
        WillReturnError(errors.New("db error"))

    _, err := repo.GetSceneByID(id)
    if err == nil {
        t.Fatal("expected error, got nil")
    }
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Error(err)
    }
}

func TestGetSceneByID_LayersQueryError(t *testing.T) {
    repo, mock := getRepositoryWithMockDB(t)
    id := uuid.New()

    mock.ExpectQuery(getSceneByIDQuery).
        WithArgs(id).
        WillReturnRows(sqlmock.NewRows(sceneColumns).
            AddRow(id, strPtr("Scene"), strPtr("Author"), pqArrayValue(t, []string{}), tables.ModerationStatusApproved))

    mock.ExpectQuery(getLayersBySceneIDQuery).
        WithArgs(id).
        WillReturnError(errors.New("layers db error"))

    _, err := repo.GetSceneByID(id)
    if err == nil {
        t.Fatal("expected error, got nil")
    }
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Error(err)
    }
}

// ---- AddScene ----

const insertSceneQuery = `INSERT INTO public."Scenes" ("ID", "Name", "Author", "UniqueTileIDs", "ModerationStatus") VALUES ($1, $2, $3, $4, $5)`
const insertLayerQuery = `INSERT INTO public."Layers" ("ID", "SceneId") VALUES ($1, $2)`
const insertTileQuery = `INSERT INTO public."Tiles" ("TileId", "Rotation", "XPos", "YPos", "LayerId") VALUES ($1, $2, $3, $4, $5)`

func TestAddScene_Success(t *testing.T) {
    repo, mock := getRepositoryWithMockDB(t)

    request := model.Scene{
        Name:   "New Scene",
        Author: "Author",
        Layers: []model.Layer{
            {Tiles: []model.Tile{
                {TileID: "tile-1", Rotation: 90, XPos: 3, YPos: 5},
            }},
        },
    }

    mock.ExpectBegin()
    mock.ExpectPrepare(insertSceneQuery).
        ExpectExec().
        WithArgs(sqlmock.AnyArg(), "New Scene", "Author", sqlmock.AnyArg(), tables.ModerationStatusPending).
        WillReturnResult(sqlmock.NewResult(1, 1))
    mock.ExpectPrepare(insertLayerQuery).
        ExpectExec().
        WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
        WillReturnResult(sqlmock.NewResult(1, 1))
    mock.ExpectPrepare(insertTileQuery).
        ExpectExec().
        WithArgs("tile-1", 90, 3, 5, sqlmock.AnyArg()).
        WillReturnResult(sqlmock.NewResult(1, 1))
    mock.ExpectCommit()

    if err := repo.AddScene(request); err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Error(err)
    }
}

func TestAddScene_NoLayers_SkipsLayerAndTileInsert(t *testing.T) {
    repo, mock := getRepositoryWithMockDB(t)

    request := model.Scene{
        Name:   "Scene No Layers",
        Author: "Author",
        Layers: []model.Layer{},
    }

    mock.ExpectBegin()
    mock.ExpectPrepare(insertSceneQuery).
        ExpectExec().
        WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
        WillReturnResult(sqlmock.NewResult(1, 1))
    mock.ExpectPrepare(insertLayerQuery) // prepared but no exec since Layers is empty
    // No tile prepare/exec expected since Layers is empty.
    mock.ExpectCommit()

    if err := repo.AddScene(request); err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Error(err)
    }
}

func TestAddScene_BeginError(t *testing.T) {
    repo, mock := getRepositoryWithMockDB(t)

    mock.ExpectBegin().WillReturnError(errors.New("begin failed"))

    if err := repo.AddScene(model.Scene{Name: "S", Author: "A"}); err == nil {
        t.Fatal("expected error, got nil")
    }
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Error(err)
    }
}

func TestAddScene_InsertSceneError_RollsBack(t *testing.T) {
    repo, mock := getRepositoryWithMockDB(t)

    mock.ExpectBegin()
    mock.ExpectPrepare(insertSceneQuery).
        ExpectExec().
        WillReturnError(errors.New("insert scene failed"))
    mock.ExpectRollback()

    if err := repo.AddScene(model.Scene{Name: "S", Author: "A", Layers: []model.Layer{}}); err == nil {
        t.Fatal("expected error, got nil")
    }
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Error(err)
    }
}

func TestAddScene_InsertTilesError_RollsBack(t *testing.T) {
    repo, mock := getRepositoryWithMockDB(t)

    request := model.Scene{
        Name:   "S",
        Author: "A",
        Layers: []model.Layer{
            {Tiles: []model.Tile{{TileID: "tile-1"}}},
        },
    }

    mock.ExpectBegin()
    mock.ExpectPrepare(insertSceneQuery).
        ExpectExec().
        WillReturnResult(sqlmock.NewResult(1, 1))
    mock.ExpectPrepare(insertLayerQuery).
        ExpectExec().
        WillReturnResult(sqlmock.NewResult(1, 1))
    mock.ExpectPrepare(insertTileQuery).
        ExpectExec().
        WillReturnError(errors.New("insert tile failed"))
    mock.ExpectRollback()

    if err := repo.AddScene(request); err == nil {
        t.Fatal("expected error, got nil")
    }
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Error(err)
    }
}