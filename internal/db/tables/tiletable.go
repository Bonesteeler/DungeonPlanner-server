package tables

import (
	  "database/sql"
		"fmt"

    "github.com/google/uuid"
)

type Tile struct {
    ID       uuid.UUID
    TileId   *string
    Rotation int
    XPos     int
    YPos     int
		SceneId  uuid.UUID
}

func GetTilesBySceneID(db *sql.DB, sceneID uuid.UUID) ([]Tile, error) {
		rows, err := db.Query(`SELECT "TileId", "Rotation", "XPos", "YPos", "SceneId" FROM public."Tiles" WHERE "SceneId" = $1`, sceneID)
		if err != nil {
			return nil, fmt.Errorf("failed to query tiles: %w", err)
		}
		defer rows.Close()

		var tiles []Tile
		for rows.Next() {
			var tile Tile
			if err := rows.Scan(&tile.TileId, &tile.Rotation, &tile.XPos, &tile.YPos, &tile.SceneId); err != nil {
				return nil, fmt.Errorf("failed to scan tile row: %w", err)
			}
			tiles = append(tiles, tile)
		}
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating over tile rows: %w", err)
		}
		return tiles, nil
}

func GetTilesBySceneIDs(db *sql.DB, sceneIDs []uuid.UUID) (map[uuid.UUID][]Tile, error) {
		rows, err := db.Query(`SELECT "TileId", "Rotation", "XPos", "YPos", "SceneId" FROM public."Tiles" WHERE "SceneId" IN ($1)`, sceneIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to query tiles: %w", err)
		}
		defer rows.Close()

		tiles := make(map[uuid.UUID][]Tile)
		for rows.Next() {
			var tile Tile
			if err := rows.Scan(&tile.TileId, &tile.Rotation, &tile.XPos, &tile.YPos, &tile.SceneId); err != nil {
				return nil, fmt.Errorf("failed to scan tile row: %w", err)
			}
			tiles[tile.SceneId] = append(tiles[tile.SceneId], tile)
		}
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating over tile rows: %w", err)
		}
		return tiles, nil
}

func InsertTiles(tx *sql.Tx, tiles []Tile) error {
    if len(tiles) == 0 {
        return nil // No tiles to insert, so we can return early.
    }
    stmt, err := tx.Prepare(`INSERT INTO public."Tiles" ("TileId", "Rotation", "XPos", "YPos", "SceneId") VALUES ($1, $2, $3, $4, $5)`)
    if err != nil {
        return fmt.Errorf("failed to prepare statement: %w", err)
    }
    defer stmt.Close()
    for _, tile := range tiles {
        _, err := stmt.Exec(tile.TileId, tile.Rotation, tile.XPos, tile.YPos, tile.SceneId)
        if err != nil {
            return fmt.Errorf("failed to execute statement: %w", err)
        }
    }
    return nil
}