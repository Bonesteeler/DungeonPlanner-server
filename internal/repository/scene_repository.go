package repository

import (
    "database/sql"
    "os"
    "strings"

    "github.com/google/uuid"

    "DungeonPlannerServer/internal/db"
    "DungeonPlannerServer/internal/db/tables"
)

type SceneRepository struct {
    db *sql.DB
}

func NewSceneRepository() (*SceneRepository, error) {
    conn, err := _establishConnection()
    if err != nil {
        return nil, err
    }
    return &SceneRepository{db: conn}, nil
}

func NewSceneRepositoryWithDB(db *sql.DB) *SceneRepository {
		return &SceneRepository{db: db}
}

func (r *SceneRepository) GetApprovedSceneCount() (int, error) {
    return tables.GetApprovedSceneCount(r.db)
}

func (r *SceneRepository) ListApprovedScenes(offset int) ([]tables.Scene, error) {
    return tables.ListApprovedScenes(r.db, offset, 20)
}

func (r *SceneRepository) GetSceneByID(id uuid.UUID) (*tables.Scene, error) {
    scene, err := tables.GetSceneByID(r.db, id)
    if err != nil || scene == nil {
        return nil, err
    }
    tiles, err := tables.GetTilesBySceneID(r.db, id)
    if err != nil {
        return nil, err
    }
    scene.Tiles = tiles
    return scene, nil
}

func (r *SceneRepository) AddScene(scene tables.Scene) error {
    tx, err := r.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()
    if err = tables.InsertScene(tx, scene); err != nil {
        return err
    }
    if err = tables.InsertTiles(tx, scene.Tiles); err != nil {
        return err
    }
    return tx.Commit()
}

func _establishConnection() (*sql.DB, error) {
    var password string
    if passwordBytes, err := os.ReadFile("/run/secrets/db-password"); err == nil {
        password = strings.TrimSpace(string(passwordBytes))
    }
    return db.Connect(password)
}