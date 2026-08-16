package tables

import (
	"database/sql"
	"github.com/google/uuid"
)

func StoreRoleForUser(db *sql.DB, userId uuid.UUID, role int) error {
	stmt, err := db.Prepare(`INSERT INTO public."UserRoles" ("UserID", "Role") VALUES ($1, $2)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(userId, role)
	if err != nil {
		return err
	}
	return nil
}

func GetRoleForUser(db *sql.DB, userId uuid.UUID) (int, error) {
	var role int
	err := db.QueryRow(`SELECT "Role" FROM public."UserRoles" WHERE "UserID" = $1`, userId).Scan(&role)
	if err != nil {
		return -1, err
	}
	return role, nil
}