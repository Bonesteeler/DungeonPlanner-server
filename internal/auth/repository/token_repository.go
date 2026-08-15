package repo

import "sync"

type InMemoryTokenStore struct {
	mu sync.Mutex
	tokens map[string]string
}

func NewInMemoryTokenStore() *InMemoryTokenStore {
	return &InMemoryTokenStore{
		tokens: make(map[string]string),
	}
}

func (store *InMemoryTokenStore) StoreToken(token string, userID string) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.tokens[token] = userID
}

func (store *InMemoryTokenStore) RemoveToken(token string) {
	store.mu.Lock()
	defer store.mu.Unlock()
	delete(store.tokens, token)
}

func (store *InMemoryTokenStore) IsTokenPresent(token string, userID string) bool {
	store.mu.Lock()
	defer store.mu.Unlock()
	storedUserID, exists := store.tokens[token]
	return exists && storedUserID == userID
}