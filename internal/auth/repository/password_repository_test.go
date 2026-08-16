package repo

import (
	"testing"
	"regexp"

	"github.com/DATA-DOG/go-sqlmock"
)

var passwordColumns = []string{"ID", "PasswordHash", "Email"}

func getRepositoryWithMockDB(t *testing.T) (*PasswordRepository, sqlmock.Sqlmock) {
	t.Helper()
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %s", err)
	}
	repo := NewPasswordRepositoryWithDB(mockDB)
	t.Cleanup(func() {
		mockDB.Close()
	})
	return repo, mock
}

const getPasswordHashByEmailQuery = `SELECT "PasswordHash" FROM public."Passwords" WHERE "Email" = $1`

func TestGetPasswordHashByEmail_Success(t *testing.T) {
	repo, mock := getRepositoryWithMockDB(t)
	expectedPasswordHash := "hashed_password"
	expectedEmail := "test@email.com"
	rows := sqlmock.NewRows([]string{"PasswordHash"}).AddRow(expectedPasswordHash)

	mock.ExpectQuery(regexp.QuoteMeta(getPasswordHashByEmailQuery)).WithArgs(expectedEmail).WillReturnRows(rows)

	passwordHash, err := repo.GetPasswordHashByEmail(expectedEmail)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if passwordHash != expectedPasswordHash {
		t.Fatalf("expected password hash %v, got %v", expectedPasswordHash, passwordHash)
	}
}