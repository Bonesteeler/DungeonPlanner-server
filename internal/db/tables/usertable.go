package tables

import (
		"database/sql"
		"github.com/google/uuid"
)

func IsEmailExists(db *sql.DB, email string) (bool, error) {
	var exists bool
	err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM public."Users" WHERE "Email" = $1)`, email).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func GetUserIdByEmail(db *sql.DB, email string) (uuid.UUID, error) {
	var userId uuid.UUID
	err := db.QueryRow(`SELECT "ID" FROM public."Users" WHERE "Email" = $1`, email).Scan(&userId)
	if err != nil {
		return uuid.Nil, err
	}
	return userId, nil
}

func AddUser(db *sql.DB, id uuid.UUID, username string, email string) error {
	stmt, err := db.Prepare(`INSERT INTO public."Users" ("ID", "Username", "Email") VALUES ($1, $2, $3)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(id, username, email)
	if err != nil {
		return err
	}
	return nil
}