package tables

import (
    "database/sql"
    "fmt"

    "github.com/google/uuid"
    "github.com/lib/pq"
)

type ModerationStatus int

const (
		ModerationStatusPending ModerationStatus = iota
		ModerationStatusApproved
		ModerationStatusRejected
)

type Scene struct {
    ID               uuid.UUID
    Name             *string
    Author           *string
    UniqueTileIDs    []string
		Tiles 					 []Tile
    ModerationStatus ModerationStatus
}

func GetAllScenes(db *sql.DB) ([]Scene, error) {
    rows, err := db.Query(`SELECT "ID", "Name", "Author", "UniqueTileIDs", "ModerationStatus" FROM public."Scenes"`)
    if err != nil {
        return nil, fmt.Errorf("failed to query scenes: %w", err)
    }
    defer rows.Close()

    var scenes []Scene
    for rows.Next() {
        var s Scene
        err := rows.Scan(
            &s.ID,
            &s.Name,
            &s.Author,
            pq.Array(&s.UniqueTileIDs),
            &s.ModerationStatus,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to scan scene row: %w", err)
        }
        scenes = append(scenes, s)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating scene rows: %w", err)
    }

    return scenes, nil
}

func GetSceneByID(db *sql.DB, id uuid.UUID) (*Scene, error) {
    row := db.QueryRow(`SELECT "ID", "Name", "Author", "UniqueTileIDs", "ModerationStatus" FROM public."Scenes" WHERE "ID" = $1`, id)

    var s Scene
    err := row.Scan(
        &s.ID,
        &s.Name,
        &s.Author,
        pq.Array(&s.UniqueTileIDs),
        &s.ModerationStatus,
    )
    if err == sql.ErrNoRows {
        return nil, nil
    }
    if err != nil {
        return nil, fmt.Errorf("failed to scan scene row: %w", err)
    }

    return &s, nil
}

func GetSceneCountsByModerationStatus(db *sql.DB) (map[ModerationStatus]int, error) {
		rows, err := db.Query(`SELECT "ModerationStatus", COUNT(*) FROM public."Scenes" GROUP BY "ModerationStatus"`)
		if err != nil {
			return nil, fmt.Errorf("failed to query scene counts: %w", err)
		}
		defer rows.Close()
		counts := make(map[ModerationStatus]int)
		for rows.Next() {
			var status ModerationStatus
			var count int
			if err := rows.Scan(&status, &count); err != nil {
				return nil, fmt.Errorf("failed to scan scene count row: %w", err)
			}
			counts[status] = count
		}
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating scene count rows: %w", err)
		}
		return counts, nil
}

func GetApprovedSceneCount(db *sql.DB) (int, error) {
		var count int
		err := db.QueryRow(`SELECT COUNT(*) FROM public."Scenes" WHERE "ModerationStatus" = $1`, ModerationStatusApproved).Scan(&count)
		if err != nil {
			return 0, fmt.Errorf("failed to query approved scene count: %w", err)
		}
		return count, nil
}

func ListApprovedScenes(db *sql.DB, offset, limit int) ([]Scene, error) {
		rows, err := db.Query(`SELECT "ID", "Name", "Author", "UniqueTileIDs", "ModerationStatus" FROM public."Scenes" WHERE "ModerationStatus" = $1 ORDER BY "ID" OFFSET $2 LIMIT $3`, ModerationStatusApproved, offset, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to query approved scenes: %w", err)
		}
		defer rows.Close()
		var scenes []Scene
		for rows.Next() {
			var s Scene
			err := rows.Scan(
				&s.ID,
				&s.Name,
				&s.Author,
				pq.Array(&s.UniqueTileIDs),
				&s.ModerationStatus,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to scan approved scene row: %w", err)
			}
			scenes = append(scenes, s)
		}
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating approved scene rows: %w", err)
		}
		return scenes, nil
}

func InsertScene(tx *sql.Tx, scene Scene) error {
		stmt, err := tx.Prepare(`INSERT INTO public."Scenes" ("ID", "Name", "Author", "UniqueTileIDs", "ModerationStatus") VALUES ($1, $2, $3, $4, $5)`)
		if err != nil {
			return fmt.Errorf("failed to prepare statement: %w", err)
		}
		defer stmt.Close()
		_, err = stmt.Exec(scene.ID, scene.Name, scene.Author, pq.Array(scene.UniqueTileIDs), scene.ModerationStatus)
		if err != nil {
			return fmt.Errorf("failed to execute statement: %w", err)
		}
		return nil
}
