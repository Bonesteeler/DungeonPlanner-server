package db

import (
  "database/sql"
  "fmt"
	"os"

	// Postgres driver
  _ "github.com/lib/pq"
)

func Connect(password string) (*sql.DB, error) {
		host := getEnv("DB_HOST", "localhost")
    port := getEnv("DB_PORT", "5432")
    user := getEnv("DB_USER", "postgres")
    dbname := getEnv("DB_NAME", "scenes")

    psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
        host, port, user, string(password), dbname)
  db, err := sql.Open("postgres", psqlInfo)
  if err != nil {
    return nil, err
  }

  err = db.Ping()
  if err != nil {
    return nil, err
  }

  fmt.Println("Successfully connected!")
  return db, nil
}

func GetTableNames(db *sql.DB) ([]string, error) {
		rows, err := db.Query("SELECT table_name FROM information_schema.tables WHERE table_schema='public'")
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var tableNames []string
		for rows.Next() {
			var tableName string
			if err := rows.Scan(&tableName); err != nil {
				return nil, err
			}
			tableNames = append(tableNames, tableName)
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return tableNames, nil
}

func getEnv(key, fallback string) string {
    if val := os.Getenv(key); val != "" {
        return val
    }
    return fallback
}