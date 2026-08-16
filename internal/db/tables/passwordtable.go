package tables

import (
		"database/sql"

		"github.com/google/uuid"
)

type UserCredentials struct {
		ID           uuid.UUID
		PasswordHash string
		Email        string
}

func GetPasswordHashByEmail(db *sql.DB, email string) (string, error) {
		var passwordHash string
		err := db.QueryRow(`SELECT "PasswordHash" FROM public."Passwords" WHERE "Email" = $1`, email).Scan(&passwordHash)
		if err != nil {
			return "", err
		}
		return passwordHash, nil
}

func StorePasswordHash(db *sql.DB, id uuid.UUID, passwordHash string, email string) error {
	stmt, err := db.Prepare(`INSERT INTO public."Passwords" ("ID", "PasswordHash", "Email") VALUES ($1, $2, $3)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(id, passwordHash, email)
	if err != nil {
		return err
	}
	return nil
}

func IsEmailExists(db *sql.DB, email string) (bool, error) {
	var exists bool
	err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM public."Passwords" WHERE "Email" = $1)`, email).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}