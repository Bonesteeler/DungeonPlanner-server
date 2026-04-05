package model

type Scene struct {
	ID            string
	Name          string
	Author        string
	UniqueTileIDs []string
	Layers        []Layer
}

type Layer struct {
	Height int
	Tiles  []Tile
}

type Tile struct {
	TileID   string
	Rotation int
	XPos     int
	YPos     int
}