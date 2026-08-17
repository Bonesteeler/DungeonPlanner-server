package tables

func GetRoleNameFromId(db DBTX, roleId int) (string, error) {
	var roleName string
	err := db.QueryRow(`SELECT "Name" FROM public."Roles" WHERE "ID" = $1`, roleId).Scan(&roleName)
	if err != nil {
		return "", err
	}
	return roleName, nil
}
