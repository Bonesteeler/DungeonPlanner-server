package repo
import (
	"database/sql"
	"errors"
	"github.com/google/uuid"

	"github.com/labstack/echo/v4"

	"DungeonPlannerServer/internal/db"
	"DungeonPlannerServer/internal/db/tables"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(e *echo.Echo) (*UserRepository, error) {
		password, err := db.GetPasswordSecret()
		if err != nil {
			return nil, err
		}
		connection, err := db.Connect(password, e)
		if err != nil {
			return nil, err
		}
		return &UserRepository{db: connection}, nil
}

func NewUserRepositoryWithDB(db *sql.DB) *UserRepository {
		return &UserRepository{db: db}
}

func (r *UserRepository) GetPasswordHashByUserId(id uuid.UUID) (string, error) {
	return tables.GetPasswordHashByUserId(r.db, id)
}

func (r *UserRepository) StorePasswordHash(id uuid.UUID, passwordHash string) error {
	return tables.StorePasswordHash(r.db, id, passwordHash)
}

func (r *UserRepository) IsEmailExists(email string) (bool, error) {
	return tables.IsEmailExists(r.db, email)
}

func (r *UserRepository) GetUserIdByEmail(email string) (uuid.UUID, error) {
	return tables.GetUserIdByEmail(r.db, email)
}

func (r *UserRepository) GetUserRoleByUserId(id uuid.UUID) (string, error) {
	roleId, err := tables.GetRoleForUser(r.db, id)
	if err != nil {
		return "", err
	}
	roleName, err := tables.GetRoleNameFromId(r.db, roleId)
	if err != nil {
		return "", err
	}
	return roleName, nil
}

func (r *UserRepository) AddUser(id uuid.UUID, username string, email string, passwordHash string) error {
	tx , err := r.db.Begin()
	if err != nil {
		return err
	}
	emailExists, err := tables.IsEmailExists(tx, email)
	if err != nil {
		tx.Rollback()
		return err
	}
	if emailExists {
		tx.Rollback()
		return errors.New("email already exists")
	}
	err = tables.AddUser(tx, id, username, email)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tables.StorePasswordHash(tx, id, passwordHash)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tables.StoreRoleForUser(tx, id, 1)
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}