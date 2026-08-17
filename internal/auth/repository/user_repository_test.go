package repo

import (
	"database/sql"
	"testing"
	"regexp"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

var passwordColumns = []string{"ID", "PasswordHash"}

const getPasswordHashByUserIdQuery = `SELECT "PasswordHash" FROM public."Passwords" WHERE "ID" = $1`
const storePasswordHashQuery = `INSERT INTO public."Passwords" ("ID", "PasswordHash") VALUES ($1, $2)`
const isEmailExistsQuery = `SELECT EXISTS(SELECT 1 FROM public."Users" WHERE "Email" = $1)`
const addUserQuery = `INSERT INTO public."Users" ("ID", "Username", "Email") VALUES ($1, $2, $3)`
const storeRoleForUserQuery = `INSERT INTO public."UserRoles" ("UserID", "Role") VALUES ($1, $2)`

func getRepositoryWithMockDB(t *testing.T) (*UserRepository, sqlmock.Sqlmock) {
	t.Helper()
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %s", err)
	}
	repo := NewUserRepositoryWithDB(mockDB)
	t.Cleanup(func() {
		mockDB.Close()
	})
	return repo, mock
}

func TestGetPasswordHashByUserId_Success(t *testing.T) {
	repo, mock := getRepositoryWithMockDB(t)
	expectedPasswordHash := "hashed_password"
	expectedUserId := uuid.New()
	rows := sqlmock.NewRows([]string{"PasswordHash"}).AddRow(expectedPasswordHash)

	mock.ExpectQuery(regexp.QuoteMeta(getPasswordHashByUserIdQuery)).WithArgs(expectedUserId).WillReturnRows(rows)

	passwordHash, err := repo.GetPasswordHashByUserId(expectedUserId)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if passwordHash != expectedPasswordHash {
		t.Fatalf("expected password hash %v, got %v", expectedPasswordHash, passwordHash)
	}
}

func TestStorePasswordHash_Success(t *testing.T) {
	repo, mock := getRepositoryWithMockDB(t)
	expectedUserId := uuid.New()
	expectedPasswordHash := "hashed_password"
	mock.ExpectPrepare(regexp.QuoteMeta(storePasswordHashQuery)).ExpectExec().
		WithArgs(expectedUserId, expectedPasswordHash).
		WillReturnResult(sqlmock.NewResult(1, 1))
	err := repo.StorePasswordHash(expectedUserId, expectedPasswordHash)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestIsEmailExists_Success(t *testing.T) {
	repo, mock := getRepositoryWithMockDB(t)
	expectedEmail := "test@example.com"
	rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)

	mock.ExpectQuery(regexp.QuoteMeta(isEmailExistsQuery)).
		WithArgs(expectedEmail).
		WillReturnRows(rows)

	exists, err := repo.IsEmailExists(expectedEmail)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !exists {
		t.Fatalf("expected email to exist")
	}
}

// AddUser

func TestAddUser_Success(t *testing.T) {
	repo, mock := getRepositoryWithMockDB(t)
	expectedUserId := uuid.New()
	expectedUsername := "testuser"
	expectedEmail := "test@example.com"
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(isEmailExistsQuery)).
		WithArgs(expectedEmail).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectPrepare(regexp.QuoteMeta(addUserQuery)).
	  ExpectExec().
		WithArgs(expectedUserId, expectedUsername, expectedEmail).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectPrepare(regexp.QuoteMeta(storePasswordHashQuery)).
	  ExpectExec().
		WithArgs(expectedUserId, "hashed_password").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectPrepare(regexp.QuoteMeta(storeRoleForUserQuery)).
		ExpectExec().
		WithArgs(expectedUserId, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.AddUser(expectedUserId, expectedUsername, expectedEmail, "hashed_password")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestAddUser_EmailAlreadyExists(t *testing.T) {
	repo, mock := getRepositoryWithMockDB(t)
	expectedUserId := uuid.New()
	expectedUsername := "testuser"
	expectedEmail := "test@example.com"
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(isEmailExistsQuery)).
		WithArgs(expectedEmail).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectRollback()

	err := repo.AddUser(expectedUserId, expectedUsername, expectedEmail, "hashed_password")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestAddUser_IsEmailExistsFails(t *testing.T) {
	repo, mock := getRepositoryWithMockDB(t)
	expectedUserId := uuid.New()
	expectedUsername := "testuser"
	expectedEmail := "test@example.com"
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(isEmailExistsQuery)).
		WithArgs(expectedEmail).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()
	
	err := repo.AddUser(expectedUserId, expectedUsername, expectedEmail, "hashed_password")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestAddUser_AddUserFails(t *testing.T) {
	repo, mock := getRepositoryWithMockDB(t)
	expectedUserId := uuid.New()
	expectedUsername := "testuser"
	expectedEmail := "test@example.com"
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(isEmailExistsQuery)).
		WithArgs(expectedEmail).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectPrepare(regexp.QuoteMeta(addUserQuery)).
		ExpectExec().
		WithArgs(expectedUserId, expectedUsername, expectedEmail).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	err := repo.AddUser(expectedUserId, expectedUsername, expectedEmail, "hashed_password")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestAddUser_StorePasswordHashFails(t *testing.T) {
	repo, mock := getRepositoryWithMockDB(t)
	expectedUserId := uuid.New()
	expectedUsername := "testuser"
	expectedEmail := "test@example.com"
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(isEmailExistsQuery)).
		WithArgs(expectedEmail).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectPrepare(regexp.QuoteMeta(addUserQuery)).
		ExpectExec().
		WithArgs(expectedUserId, expectedUsername, expectedEmail).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectPrepare(regexp.QuoteMeta(storePasswordHashQuery)).
		ExpectExec().
		WithArgs(expectedUserId, "hashed_password").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	err := repo.AddUser(expectedUserId, expectedUsername, expectedEmail, "hashed_password")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestAddUser_StoreRoleForUserFails(t *testing.T) {
	repo, mock := getRepositoryWithMockDB(t)
	expectedUserId := uuid.New()
	expectedUsername := "testuser"
	expectedEmail := "test@example.com"
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(isEmailExistsQuery)).
		WithArgs(expectedEmail).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectPrepare(regexp.QuoteMeta(addUserQuery)).
		ExpectExec().
		WithArgs(expectedUserId, expectedUsername, expectedEmail).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectPrepare(regexp.QuoteMeta(storePasswordHashQuery)).
		ExpectExec().
		WithArgs(expectedUserId, "hashed_password").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectPrepare(regexp.QuoteMeta(storeRoleForUserQuery)).
		ExpectExec().
		WithArgs(expectedUserId, 1).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	err := repo.AddUser(expectedUserId, expectedUsername, expectedEmail, "hashed_password")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}