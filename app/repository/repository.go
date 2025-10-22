package repository

import "time"

// KeyValueRepository defines the interface for key-value storage operations
type KeyValueRepository interface {
	// Set stores a key-value pair with optional expiration
	Set(key, value string, expiration *time.Duration) error

	// Get retrieves a value by key, returns (value, exists)
	Get(key string) (string, bool)

	// Delete removes a key from storage
	Delete(key string) error

	// Exists checks if a key exists in storage
	Exists(key string) bool

	// Keys returns all keys matching a pattern (for future use)
	Keys(pattern string) ([]string, error)

	// Clear removes all keys from storage
	Clear() error

	// Size returns the number of keys in storage
	Size() int
}
