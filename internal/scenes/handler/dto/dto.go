package dto

type TileRequest struct {
	TileID   string `json:"tile_id"`
	Rotation int    `json:"rotation"`
	XPos     int    `json:"x_pos"`
	YPos     int    `json:"y_pos"`
}

type TileResponse struct {
	TileID   string `json:"tile_id"`
	Rotation int    `json:"rotation"`
	XPos     int    `json:"x_pos"`
	YPos     int    `json:"y_pos"`
}

type LayerRequest struct {
	Height int           `json:"height"`
	Tiles  []TileRequest `json:"tiles"`
}

type LayerResponse struct {
	Height int            `json:"height"`
	Tiles  []TileResponse `json:"tiles"`
}

type AddSceneRequest struct {
	Name   string         `json:"name"`
	Author string         `json:"author"`
	Layers []LayerRequest `json:"layers"`
}

type SceneResponse struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	Author        string          `json:"author"`
	UniqueTileIDs []string        `json:"unique_tile_ids"`
	Layers        []LayerResponse `json:"layers"`
}

type SceneListResponse struct {
	Scenes     []SceneResponse `json:"scenes"`
	TotalCount int             `json:"total_count"`
	PageSize   int             `json:"page_size"`
}

type SceneStatsResponse struct {
	ApprovedScenes int `json:"approved_scenes"`
}