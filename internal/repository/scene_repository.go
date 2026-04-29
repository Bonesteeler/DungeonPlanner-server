package repository

import (
	"database/sql"
	"os"
	"strings"

	"github.com/google/uuid"

	"DungeonPlannerServer/internal/db"
	"DungeonPlannerServer/internal/db/tables"
	"DungeonPlannerServer/internal/model"
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

func (r *SceneRepository) ListApprovedScenes(offset int) ([]model.Scene, error) {
    scenes, err := tables.ListApprovedScenes(r.db, offset, 20)
    if err != nil {
        return nil, err
    }
    response := make([]model.Scene, 0, len(scenes))
    for _, s := range scenes {
        response = append(response, model.Scene{
            ID:            s.ID.String(),
            Name:          derefString(s.Name),
            Author:        derefString(s.Author),
            UniqueTileIDs: s.UniqueTileIDs,
        })
    }
    return response, nil
}

func (r *SceneRepository) GetSceneByID(id uuid.UUID) (*model.Scene, error) {
    scene, err := tables.GetSceneByID(r.db, id)
    if err != nil || scene == nil {
        return nil, err
    }
    layers, err := tables.GetLayersBySceneID(r.db, id)
    if err != nil {
        return nil, err
    }
    response := &model.Scene{
        ID:            scene.ID.String(),
        Name:          derefString(scene.Name),
        Author:        derefString(scene.Author),
        UniqueTileIDs: scene.UniqueTileIDs,
    }
    for _, layer := range layers {
        tiles, err := tables.GetTilesByLayerID(r.db, layer.ID)
        if err != nil {
            return nil, err
        }
        tileResponses := make([]model.Tile, 0, len(tiles))
        for _, t := range tiles {
            tileResponses = append(tileResponses, model.Tile{
                TileID:   derefString(t.TileId),
                Rotation: t.Rotation,
                XPos:     t.XPos,
                YPos:     t.YPos,
            })
        }
        response.Layers = append(response.Layers, model.Layer{
            Tiles: tileResponses,
        })
    }
    return response, nil
}

func (r *SceneRepository) AddScene(request model.Scene) error {
    uniqueTileIDMap := make(map[string]bool)
    for _, layer := range request.Layers {
        for _, tile := range layer.Tiles {
            uniqueTileIDMap[tile.TileID] = true
        }
    }
    uniqueTileIDs := make([]string, 0, len(uniqueTileIDMap))
    for tileID := range uniqueTileIDMap {
        uniqueTileIDs = append(uniqueTileIDs, tileID)
    }
    sceneID := uuid.New()
    scene := tables.Scene{
        ID:            sceneID,
        Name:          &request.Name,
        Author:        &request.Author,
        UniqueTileIDs: uniqueTileIDs,
    }
    layers := make([]tables.Layer, 0, len(request.Layers))
    for _, lr := range request.Layers {
        layers = append(layers, tables.Layer{
            ID:      uuid.New(),
            SceneId: sceneID,
        })
        _ = lr
    }

    tx, err := r.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()
    if err = tables.InsertScene(tx, scene); err != nil {
        return err
    }
    if err = tables.InsertLayers(tx, layers); err != nil {
        return err
    }
    for i, layer := range layers {
        tiles := make([]tables.Tile, 0, len(request.Layers[i].Tiles))
        for _, tr := range request.Layers[i].Tiles {
            tileID := tr.TileID
            tiles = append(tiles, tables.Tile{
                TileId:  &tileID,
                Rotation: tr.Rotation,
                XPos:    tr.XPos,
                YPos:    tr.YPos,
                LayerId: layer.ID,
            })
        }
        if err = tables.InsertTiles(tx, tiles); err != nil {
            return err
        }
    }
    return tx.Commit()
}

func derefString(s *string) string {
    if s == nil {
        return ""
    }
    return *s
}

func _establishConnection() (*sql.DB, error) {
    var password string
    if passwordBytes, err := os.ReadFile("/run/secrets/db-password"); err == nil {
        password = strings.TrimSpace(string(passwordBytes))
    }
    return db.Connect(password)
}