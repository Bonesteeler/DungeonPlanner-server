package dto

type TileRequest struct {
	TileID   string
	Rotation int
	XPos     int
	YPos     int
}

type AddSceneRequest struct {
	Name   string
	Author string
	Tiles  []TileRequest
}

type TileResponse struct {
	TileID   string
	Rotation int
	XPos     int
	YPos     int
}

type SceneResponse struct {
	ID            string
	Name          string
	Author        string
	UniqueTileIDs []string
	Tiles         []TileResponse
}

type SceneStatsResponse struct {
	ApprovedScenes int
}