package db

import (
  "database/sql"
  "fmt"
	"os"
	"github.com/labstack/echo/v4"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	// Postgres driver
  _ "github.com/lib/pq"
)

func Connect(password string, e *echo.Echo) (*sql.DB, error) {
		var db *sql.DB 
		var err error
	  if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
        db, err = _connectToUrl(dbURL)
				if err != nil {
					return nil, fmt.Errorf("failed to connect using DATABASE_URL: %w", err)
				}
    } else {
			db, err = _connectWithParams(password)
			if err != nil {
				return nil, fmt.Errorf("failed to connect using parameters: %w", err)
			}
		}

		fmt.Println("Successfully connected!")
		if err = _runMigrations(db, e); err != nil {
			return nil, err
		}
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

func _connectToUrl(url string) (*sql.DB, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
			return nil, err
	}
	return db, db.Ping()
}

func _connectWithParams(password string) (*sql.DB, error) {
		host := getEnv("DB_HOST", "localhost")
    port := getEnv("DB_PORT", "5432")
    user := getEnv("DB_USER", "postgres")
    dbname := getEnv("DB_NAME", "scenes")

    psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=prefer",
        host, port, user, string(password), dbname)
		db, err := sql.Open("postgres", psqlInfo)
		if err != nil {
			return nil, err
		}

		return db, db.Ping()
}

func _runMigrations(dbDSN *sql.DB, e *echo.Echo) error {
	driver, err := postgres.WithInstance(dbDSN, &postgres.Config{})
	 if err != nil {
			 return fmt.Errorf("failed to create migration driver: %v", err)
	 }
   m, err := migrate.NewWithDatabaseInstance(
       "file://./internal/db/migrations",
       "postgres", driver)
   if err != nil {
       return fmt.Errorf("failed to initialize migrations: %v", err)
   }

   if err := m.Up(); err != nil && err != migrate.ErrNoChange {
       return fmt.Errorf("failed to apply migrations: %v", err)
   }

   e.Logger.Info("Migrations applied successfully!")
   return nil
}