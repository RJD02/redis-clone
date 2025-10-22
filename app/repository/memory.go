package repository

import (
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/storage"
)

// MemoryRepository implements KeyValueRepository using in-memory storage
type MemoryRepository struct {
	storage *storage.ExpiringDict
}

// NewMemoryRepository creates a new memory-based repository
func NewMemoryRepository() KeyValueRepository {
	return &MemoryRepository{
		storage: storage.NewExpiringDict(),
	}
}

// NewMemoryRepositoryWithStorage creates a repository with existing storage
func NewMemoryRepositoryWithStorage(dict *storage.ExpiringDict) KeyValueRepository {
	return &MemoryRepository{
		storage: dict,
	}
}

// Set stores a key-value pair with optional expiration
func (r *MemoryRepository) Set(key, value string, expiration *time.Duration) error {
	r.storage.Set(key, value, expiration)
	return nil
}

// Get retrieves a value by key, returns (value, exists)
func (r *MemoryRepository) Get(key string) (string, bool) {
	return r.storage.Get(key)
}

// Delete removes a key from storage
func (r *MemoryRepository) Delete(key string) error {
	r.storage.Delete(key)
	return nil
}

// Exists checks if a key exists in storage
func (r *MemoryRepository) Exists(key string) bool {
	_, exists := r.storage.Get(key)
	return exists
}

// Keys returns all keys matching a pattern (simplified implementation)
func (r *MemoryRepository) Keys(pattern string) ([]string, error) {
	// For now, return empty slice - this would need proper implementation
	// with pattern matching in a production system
	return []string{}, nil
}

// Clear removes all keys from storage
func (r *MemoryRepository) Clear() error {
	// Create a new dictionary to clear all data
	r.storage = storage.NewExpiringDict()
	return nil
}

// Size returns the number of keys in storage
func (r *MemoryRepository) Size() int {
	// This would need to be implemented in ExpiringDict
	// For now, return 0 as a placeholder
	return 0
}
