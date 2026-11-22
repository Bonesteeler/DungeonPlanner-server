package repository

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/lib/pq"

	"dungeonplannerserver/dto"
	"dungeonplannerserver/internal/scenes/model"
)

var sceneColumns = []string{"ID", "Name", "Author", "Username", "UniqueTileIDs", "ModerationStatus"}
var layerColumns = []string{"ID", "SceneId", "Height"}
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

func uuidPtr(u uuid.UUID) *uuid.UUID { return &u }

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
        WithArgs(dto.ModerationStatusApproved).
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
        WithArgs(dto.ModerationStatusApproved).
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

const listApprovedQuery = `SELECT s."ID", s."Name", s."Author", u."Username", s."UniqueTileIDs", s."ModerationStatus" FROM public."Scenes" s LEFT JOIN public."Users" u ON u."ID" = s."Author" WHERE s."ModerationStatus" = $1 ORDER BY s."ID" OFFSET $2 LIMIT $3`

func TestListApprovedScenes_ReturnsScenes(t *testing.T) {
    repo, mock := getRepositoryWithMockDB(t)

    id := uuid.New()
    authorID := uuid.New()
    rows := sqlmock.NewRows(sceneColumns).
        AddRow(id, strPtr("Test Scene"), uuidPtr(authorID), strPtr("Author"), pqArrayValue(t, []string{"tile1", "tile2"}), dto.ModerationStatusApproved)

    mock.ExpectQuery(listApprovedQuery).
        WithArgs(dto.ModerationStatusApproved, 0, 20).
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
        WithArgs(dto.ModerationStatusApproved, 0, 20).
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
        WithArgs(dto.ModerationStatusApproved, 40, 20).
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

const getSceneByIDQuery = `SELECT s."ID", s."Name", s."Author", u."Username", s."UniqueTileIDs", s."ModerationStatus" FROM public."Scenes" s LEFT JOIN public."Users" u ON u."ID" = s."Author" WHERE s."ID" = $1`
const getLayersBySceneIDQuery = `SELECT "ID", "SceneId", "Height" FROM public."Layers" WHERE "SceneId" = $1`
const getTilesByLayerIDQuery = `SELECT "TileId", "Rotation", "XPos", "YPos", "LayerId" FROM public."Tiles" WHERE "LayerId" = $1`

func TestGetSceneByID_Found(t *testing.T) {
    repo, mock := getRepositoryWithMockDB(t)

    sceneID := uuid.New()
    layerID := uuid.New()
    tileID := "tile-a"

    mock.ExpectQuery(getSceneByIDQuery).
        WithArgs(sceneID).
        WillReturnRows(sqlmock.NewRows(sceneColumns).
            AddRow(sceneID, strPtr("My Scene"), uuidPtr(uuid.New()), strPtr("Author"), pqArrayValue(t, []string{"tile-a"}), dto.ModerationStatusApproved))

    mock.ExpectQuery(getLayersBySceneIDQuery).
        WithArgs(sceneID).
        WillReturnRows(sqlmock.NewRows(layerColumns).
            AddRow(layerID, sceneID, 0))

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
            AddRow(id, strPtr("Scene With No Layers"), uuidPtr(uuid.New()), strPtr("Author"), pqArrayValue(t, []string{}), dto.ModerationStatusApproved))
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
            AddRow(id, strPtr("Scene"), uuidPtr(uuid.New()), strPtr("Author"), pqArrayValue(t, []string{}), dto.ModerationStatusApproved))

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
const insertLayerQuery = `INSERT INTO public."Layers" ("ID", "SceneId", "Height") VALUES ($1, $2, $3)`
const insertTileQuery = `INSERT INTO public."Tiles" ("TileId", "Rotation", "XPos", "YPos", "LayerId") VALUES ($1, $2, $3, $4, $5)`

func TestAddScene_Success(t *testing.T) {
    repo, mock := getRepositoryWithMockDB(t)

    sceneID := uuid.New()
    authorID := uuid.New()
    request := model.Scene{
        ID:       sceneID.String(),
        Name:     "New Scene",
        AuthorID: authorID,
        Layers: []model.Layer{
            {Tiles: []model.Tile{
                {TileID: "tile-1", Rotation: 90, XPos: 3, YPos: 5},
            }},
        },
    }

    mock.ExpectBegin()
    mock.ExpectPrepare(insertSceneQuery).
        ExpectExec().
        WithArgs(sceneID, "New Scene", authorID, sqlmock.AnyArg(), dto.ModerationStatusPending).
        WillReturnResult(sqlmock.NewResult(1, 1))
    mock.ExpectPrepare(insertLayerQuery).
        ExpectExec().
        WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
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
        ID:       uuid.New().String(),
        Name:     "Scene No Layers",
        AuthorID: uuid.New(),
        Layers:   []model.Layer{},
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

    if err := repo.AddScene(model.Scene{ID: uuid.New().String(), Name: "S", AuthorID: uuid.New()}); err == nil {
        t.Fatal("expected error, got nil")
    }
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Error(err)
    }
}

func TestAddScene_InvalidID_ReturnsError(t *testing.T) {
    repo, mock := getRepositoryWithMockDB(t)

    if err := repo.AddScene(model.Scene{ID: "not-a-uuid", Name: "S", AuthorID: uuid.New()}); err == nil {
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

    if err := repo.AddScene(model.Scene{ID: uuid.New().String(), Name: "S", AuthorID: uuid.New(), Layers: []model.Layer{}}); err == nil {
        t.Fatal("expected error, got nil")
    }
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Error(err)
    }
}

func TestAddScene_InsertTilesError_RollsBack(t *testing.T) {
    repo, mock := getRepositoryWithMockDB(t)

    request := model.Scene{
        ID:       uuid.New().String(),
        Name:     "S",
        AuthorID: uuid.New(),
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