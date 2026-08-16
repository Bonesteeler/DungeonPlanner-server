package repo
import (
	"database/sql"
	"github.com/google/uuid"

	"github.com/labstack/echo/v4"

	"DungeonPlannerServer/internal/db"
	"DungeonPlannerServer/internal/db/tables"
)

type PasswordRepository struct {
	db *sql.DB
}

func NewPasswordRepository(e *echo.Echo) (*PasswordRepository, error) {
		password, err := db.GetPasswordSecret()
		if err != nil {
			return nil, err
		}
		connection, err := db.Connect(password, e)
		if err != nil {
			return nil, err
		}
		return &PasswordRepository{db: connection}, nil
}

func NewPasswordRepositoryWithDB(db *sql.DB) *PasswordRepository {
		return &PasswordRepository{db: db}
}

func (r *PasswordRepository) GetPasswordHashByEmail(email string) (string, error) {
	return tables.GetPasswordHashByEmail(r.db, email)
}

func (r *PasswordRepository) StorePasswordHash(id uuid.UUID, passwordHash string, email string) error {
	return tables.StorePasswordHash(r.db, id, passwordHash, email)
}

func (r *PasswordRepository) IsEmailExists(email string) (bool, error) {
	return tables.IsEmailExists(r.db, email)
}