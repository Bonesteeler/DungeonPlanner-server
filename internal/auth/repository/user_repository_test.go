package repo

import (
	"testing"
	"regexp"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

var passwordColumns = []string{"ID", "PasswordHash"}

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

const getPasswordHashByUserIdQuery = `SELECT "PasswordHash" FROM public."Passwords" WHERE "ID" = $1`

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