package tables

import (
		"database/sql"

		"github.com/google/uuid"
)

func GetPasswordHashByUserId(db *sql.DB, id uuid.UUID) (string, error) {
		var passwordHash string
		err := db.QueryRow(`SELECT "PasswordHash" FROM public."Passwords" WHERE "ID" = $1`, id).Scan(&passwordHash)
		if err != nil {
			return "", err
		}
		return passwordHash, nil
}

func StorePasswordHash(db *sql.DB, id uuid.UUID, passwordHash string) error {
	stmt, err := db.Prepare(`INSERT INTO public."Passwords" ("ID", "PasswordHash") VALUES ($1, $2)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(id, passwordHash)
	if err != nil {
		return err
	}
	return nil
}
