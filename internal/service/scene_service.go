package service

import (
    "github.com/google/uuid"

    "DungeonPlannerServer/internal/db/tables"
    "DungeonPlannerServer/internal/handler/dto"
)

type SceneRepo interface {
    ListApprovedScenes(offset int) ([]dto.AddSceneRequest, error)
    GetSceneByID(id uuid.UUID) (*tables.Scene, error)
    AddScene(scene tables.Scene) error
}

type SceneService struct {
    repo SceneRepo
}

func NewSceneService(repo SceneRepo) *SceneService {
    return &SceneService{repo: repo}
}

func (s *SceneService) GetSceneStats() int {
    return 0
}

func (s *SceneService) ListScenes(offset int) []dto.SceneResponse {
    scenes, err := s.repo.ListApprovedScenes(offset)
    if err != nil {
        return []dto.SceneResponse{}
    }
    var response []dto.SceneResponse
    for _, scene := range scenes {
        var layers []dto.LayerResponse
        response = append(response, dto.SceneResponse{
            ID:            scene.ID.String(),
            Name:          *scene.Name,
            Author:        *scene.Author,
            UniqueTileIDs: scene.UniqueTileIDs,
            Layers:        layers,
        })
    }
    return response
}

func (s *SceneService) GetSceneByID(id uuid.UUID) *dto.SceneResponse {
    scene, err := s.repo.GetSceneByID(id)
    if err != nil || scene == nil {
        return nil
    }
    response := dto.SceneResponse{
        ID:            scene.ID.String(),
        Name:          *scene.Name,
        Author:        *scene.Author,
        UniqueTileIDs: scene.UniqueTileIDs,
    }
    for _, layer := range scene.Layers {
				tiles := make([]dto.TileResponse, 0)
				for _, tile := range layer.Tiles {
					tiles = append(tiles, dto.TileResponse{
            TileID:   *tile.TileId,
            Rotation: tile.Rotation,
            XPos:     tile.XPos,
            YPos:     tile.YPos,
        })
    }
    response.Layers = append(response.Layers, dto.LayerResponse{
        Height: layer.Height,
        Tiles:  tiles,
    })
}
return &response
}

func (s *SceneService) AddScene(request dto.AddSceneRequest) {
    uniqueTileIDMap := make(map[string]bool)
    for _, tile := range request.Tiles {
        uniqueTileIDMap[tile.TileID] = true
    }
    uniqueTileIDs := make([]string, 0, len(uniqueTileIDMap))
    for tileID := range uniqueTileIDMap {
        uniqueTileIDs = append(uniqueTileIDs, tileID)
    }
    repositoryScene := tables.Scene{
        Name:          &request.Name,
        Author:        &request.Author,
        UniqueTileIDs: uniqueTileIDs,
        Tiles:         _convertTileRequestsToTiles(request.Tiles),
    }
    _ = s.repo.AddScene(repositoryScene)
}

func _convertTileRequestsToTiles(tileRequests []dto.TileRequest) []tables.Tile {
    var tiles []tables.Tile
    for _, tileRequest := range tileRequests {
        tiles = append(tiles, tables.Tile{
            TileId:   &tileRequest.TileID,
            Rotation: tileRequest.Rotation,
            XPos:     tileRequest.XPos,
            YPos:     tileRequest.YPos,
        })
    }
    return tiles
}