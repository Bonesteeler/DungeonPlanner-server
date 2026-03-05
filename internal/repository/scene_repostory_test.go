package repository

import (
    "errors"
    "testing"

    "github.com/DATA-DOG/go-sqlmock"
    "github.com/google/uuid"
    "github.com/lib/pq"

    "DungeonPlannerServer/internal/db/tables"
)

var sceneColumns = []string{"ID", "Name", "Author", "UniqueTileIDs", "ModerationStatus"}
var tileColumns = []string{"TileId", "Rotation", "XPos", "YPos", "SceneId"}

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
    if scenes[0].ID != id {
        t.Errorf("expected ID %v, got %v", id, scenes[0].ID)
    }
		if scenes[0].Name == nil || *scenes[0].Name != "Test Scene" {
				t.Errorf("expected Name 'Test Scene', got '%v'", scenes[0].Name)
		}
		if scenes[0].Author == nil || *scenes[0].Author != "Author" {
				t.Errorf("expected Author 'Author', got '%v'", scenes[0].Author)
		}
		if len(scenes[0].UniqueTileIDs) != 2 || scenes[0].UniqueTileIDs[0] != "tile1" || scenes[0].UniqueTileIDs[1] != "tile2" {
				t.Errorf("expected UniqueTileIDs ['tile1','tile2'], got %v", scenes[0].UniqueTileIDs)
		}
		if scenes[0].ModerationStatus != tables.ModerationStatusApproved {
				t.Errorf("expected ModerationStatus %v, got %v", tables.ModerationStatusApproved, scenes[0].ModerationStatus)
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
const getTilesBySceneIDQuery = `SELECT "TileId", "Rotation", "XPos", "YPos", "SceneId" FROM public."Tiles" WHERE "SceneId" = $1`

func TestGetSceneByID_Found(t *testing.T) {
    repo, mock := getRepositoryWithMockDB(t)

    id := uuid.New()
    tileID := "tile-a"

    mock.ExpectQuery(getSceneByIDQuery).
        WithArgs(id).
        WillReturnRows(sqlmock.NewRows(sceneColumns).
            AddRow(id, strPtr("My Scene"), strPtr("Author"), pqArrayValue(t, []string{"tile-a"}), tables.ModerationStatusApproved))

    mock.ExpectQuery(getTilesBySceneIDQuery).
        WithArgs(id).
        WillReturnRows(sqlmock.NewRows(tileColumns).
            AddRow(strPtr(tileID), 0, 1, 2, id))

    scene, err := repo.GetSceneByID(id)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if scene == nil {
        t.Fatal("expected scene, got nil")
    }
    if scene.ID != id {
        t.Errorf("expected ID %v, got %v", id, scene.ID)
    }
    if len(scene.Tiles) != 1 {
        t.Errorf("expected 1 tile, got %d", len(scene.Tiles))
    }
		if scene.Tiles[0].XPos != 1 || scene.Tiles[0].YPos != 2 {
				t.Errorf("expected tile position (1,2), got (%d,%d)", scene.Tiles[0].XPos, scene.Tiles[0].YPos)
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

func TestGetSceneByID_TilesNotFound(t *testing.T) {
		repo, mock := getRepositoryWithMockDB(t)
		id := uuid.New()
		mock.ExpectQuery(getSceneByIDQuery).
				WithArgs(id).
				WillReturnRows(sqlmock.NewRows(sceneColumns).
						AddRow(id, strPtr("Scene With No Tiles"), strPtr("Author"), pqArrayValue(t, []string{}), tables.ModerationStatusApproved))
		mock.ExpectQuery(getTilesBySceneIDQuery).
				WithArgs(id).
				WillReturnRows(sqlmock.NewRows(tileColumns))

		scene, err := repo.GetSceneByID(id)
		if err != nil {
				t.Fatalf("unexpected error: %v", err)
		}
		if scene == nil {
				t.Fatal("expected scene, got nil")
		}
		if len(scene.Tiles) != 0 {
				t.Errorf("expected 0 tiles, got %d", len(scene.Tiles))
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

func TestGetSceneByID_TilesQueryError(t *testing.T) {
    repo, mock := getRepositoryWithMockDB(t)
    id := uuid.New()

    mock.ExpectQuery(getSceneByIDQuery).
        WithArgs(id).
        WillReturnRows(sqlmock.NewRows(sceneColumns).
            AddRow(id, strPtr("Scene"), strPtr("Author"), pqArrayValue(t, []string{}), tables.ModerationStatusApproved))

    mock.ExpectQuery(getTilesBySceneIDQuery).
        WithArgs(id).
        WillReturnError(errors.New("tiles db error"))

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
const insertTileQuery = `INSERT INTO public."Tiles" ("TileId", "Rotation", "XPos", "YPos", "SceneId") VALUES ($1, $2, $3, $4, $5)`

func TestAddScene_Success(t *testing.T) {
			repo, mock := getRepositoryWithMockDB(t)

    id := uuid.New()
    tileID := "tile-1"
    scene := tables.Scene{
        ID:               id,
        Name:             strPtr("New Scene"),
        Author:           strPtr("Author"),
        UniqueTileIDs:    []string{"tile-1"},
        ModerationStatus: tables.ModerationStatusPending,
        Tiles: []tables.Tile{
            {TileId: &tileID, Rotation: 90, XPos: 3, YPos: 5, SceneId: id},
        },
    }

    mock.ExpectBegin()
    mock.ExpectPrepare(insertSceneQuery).
        ExpectExec().
        WithArgs(id, "New Scene", "Author", pqArrayValue(t, []string{"tile-1"}), tables.ModerationStatusPending).
        WillReturnResult(sqlmock.NewResult(1, 1))
    mock.ExpectPrepare(insertTileQuery).
        ExpectExec().
        WithArgs("tile-1", 90, 3, 5, id).
        WillReturnResult(sqlmock.NewResult(1, 1))
    mock.ExpectCommit()

    if err := repo.AddScene(scene); err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Error(err)
    }
}

func TestAddScene_NoTiles_SkipsTileInsert(t *testing.T) {
    repo, mock := getRepositoryWithMockDB(t)

    id := uuid.New()
    scene := tables.Scene{
        ID:               id,
        Name:             strPtr("Scene No Tiles"),
        Author:           strPtr("Author"),
        UniqueTileIDs:    []string{},
        ModerationStatus: tables.ModerationStatusPending,
        Tiles:            []tables.Tile{},
    }

    mock.ExpectBegin()
    mock.ExpectPrepare(insertSceneQuery).
        ExpectExec().
        WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
        WillReturnResult(sqlmock.NewResult(1, 1))
    // No tile prepare/exec expected since Tiles is empty.
    mock.ExpectCommit()

    if err := repo.AddScene(scene); err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Error(err)
    }
}

func TestAddScene_BeginError(t *testing.T) {
    repo, mock := getRepositoryWithMockDB(t)

    mock.ExpectBegin().WillReturnError(errors.New("begin failed"))

    scene := tables.Scene{ID: uuid.New(), Name: strPtr("S"), Author: strPtr("A")}
    if err := repo.AddScene(scene); err == nil {
        t.Fatal("expected error, got nil")
    }
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Error(err)
    }
}

func TestAddScene_InsertSceneError_RollsBack(t *testing.T) {
    repo, mock := getRepositoryWithMockDB(t)

    scene := tables.Scene{
        ID: uuid.New(), Name: strPtr("S"), Author: strPtr("A"), Tiles: []tables.Tile{},
    }

    mock.ExpectBegin()
    mock.ExpectPrepare(insertSceneQuery).
        ExpectExec().
        WillReturnError(errors.New("insert scene failed"))
    mock.ExpectRollback()

    if err := repo.AddScene(scene); err == nil {
        t.Fatal("expected error, got nil")
    }
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Error(err)
    }
}

func TestAddScene_InsertTilesError_RollsBack(t *testing.T) {
    repo, mock := getRepositoryWithMockDB(t)

    id := uuid.New()
    tileID := "tile-1"
    scene := tables.Scene{
        ID:     id,
        Name:   strPtr("S"),
        Author: strPtr("A"),
        Tiles:  []tables.Tile{{TileId: &tileID, SceneId: id}},
    }

    mock.ExpectBegin()
    mock.ExpectPrepare(insertSceneQuery).
        ExpectExec().
        WillReturnResult(sqlmock.NewResult(1, 1))
    mock.ExpectPrepare(insertTileQuery).
        ExpectExec().
        WillReturnError(errors.New("insert tile failed"))
    mock.ExpectRollback()

    if err := repo.AddScene(scene); err == nil {
        t.Fatal("expected error, got nil")
    }
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Error(err)
    }
}