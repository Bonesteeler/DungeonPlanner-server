package tables

import (
	"database/sql"
)

type DBTX interface {
    Prepare(query string) (*sql.Stmt, error)
		Query(query string, args ...any) (*sql.Rows, error)
    QueryRow(query string, args ...any) *sql.Row
    Exec(query string, args ...any) (sql.Result, error)
}
