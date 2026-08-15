package repo

import (
	"testing"
)

func TestInMemoryTokenStore_IsTokenPresent(t *testing.T) {
	store := NewInMemoryTokenStore()
	token := "test-token"
	userID := "test-user"

	if store.IsTokenPresent(token, userID) {
		t.Errorf("expected token to not be present")
	} 

	store.StoreToken(token, userID)
	if !store.IsTokenPresent(token, userID) {
		t.Errorf("expected token to be present")
	}

	store.RemoveToken(token)
	if store.IsTokenPresent(token, userID) {
		t.Errorf("expected token to be removed")
	}
} 