package tables

import (
		"database/sql"
		"fmt"

		"github.com/google/uuid"
)

type Layer struct {
		ID      uuid.UUID
		SceneId uuid.UUID
}

func GetLayersBySceneID(db *sql.DB, sceneID uuid.UUID) ([]Layer, error) {
		rows, err := db.Query(`SELECT "ID", "SceneId" FROM public."Layers" WHERE "SceneId" = $1`, sceneID)
		if err != nil {
			return nil, fmt.Errorf("failed to query layers: %w", err)
		}
		defer rows.Close()

		var layers []Layer
		for rows.Next() {
			var layer Layer
			if err := rows.Scan(&layer.ID, &layer.SceneId); err != nil {
				return nil, fmt.Errorf("failed to scan layer row: %w", err)
			}
			layers = append(layers, layer)
		}
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating over layer rows: %w", err)
		}
		return layers, nil
}

func InsertLayers(tx *sql.Tx, layers []Layer) error {
		stmt, err := tx.Prepare(`INSERT INTO public."Layers" ("ID", "SceneId") VALUES ($1, $2)`)
		if err != nil {
			return fmt.Errorf("failed to prepare insert statement for layers: %w", err)
		}
		defer stmt.Close()
		for _, layer := range layers {
			if _, err := stmt.Exec(layer.ID, layer.SceneId); err != nil {
				return fmt.Errorf("failed to execute insert statement for layer: %w", err)
			}
		}
		return nil
} 