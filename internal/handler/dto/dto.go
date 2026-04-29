package dto

type TileRequest struct {
	TileID   string
	Rotation int
	XPos     int
	YPos     int
}

type TileResponse struct {
	TileID   string
	Rotation int
	XPos     int
	YPos     int
}

type LayerRequest struct {
	Height int
	Tiles []TileRequest
}

type LayerResponse struct {
	Height int
	Tiles []TileResponse
}

type AddSceneRequest struct {
	Name   string
	Author string
	Layers []LayerRequest
}

type SceneResponse struct {
	ID            string
	Name          string
	Author        string
	UniqueTileIDs []string
	Layers        []LayerResponse
}

type SceneStatsResponse struct {
	ApprovedScenes int
}