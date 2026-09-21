package tables

import (
	"github.com/google/uuid"
)

func StoreRoleForUser(db DBTX, userId uuid.UUID, role int) error {
	stmt, err := db.Prepare(`INSERT INTO public."UserRoles" ("UserID", "RoleID") VALUES ($1, $2)`)
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

func GetRoleForUser(db DBTX, userId uuid.UUID) (int, error) {
	var role int
	err := db.QueryRow(`SELECT "RoleID" FROM public."UserRoles" WHERE "UserID" = $1`, userId).Scan(&role)
	if err != nil {
		return -1, err
	}
	return role, nil
}