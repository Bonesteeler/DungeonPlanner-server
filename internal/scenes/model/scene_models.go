package model

import "github.com/google/uuid"

type Scene struct {
	ID            string
	Name          string
	AuthorID      uuid.UUID
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