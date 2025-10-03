package storage

import (
	"sync"
	"time"
)

// KeyValue represents a value with expiration
type KeyValue struct {
	Value     string
	ExpiresAt *time.Time
}

// ExpiringDict is a thread-safe dictionary with expiration support
type ExpiringDict struct {
	data map[string]KeyValue
	mu   sync.RWMutex
}

// NewExpiringDict creates a new expiring dictionary
func NewExpiringDict() *ExpiringDict {
	return &ExpiringDict{
		data: make(map[string]KeyValue),
	}
}

// Set stores a key-value pair with optional expiration
func (ed *ExpiringDict) Set(key, value string, expiration *time.Duration) {
	ed.mu.Lock()
	defer ed.mu.Unlock()

	kv := KeyValue{Value: value}
	if expiration != nil {
		expiresAt := time.Now().Add(*expiration)
		kv.ExpiresAt = &expiresAt

		// Start a goroutine to delete the key after expiration
		go func(k string, expireTime time.Time) {
			time.Sleep(time.Until(expireTime))
			ed.Delete(k)
		}(key, expiresAt)
	}

	ed.data[key] = kv
}

// Get retrieves a value by key, checking expiration
func (ed *ExpiringDict) Get(key string) (string, bool) {
	ed.mu.RLock()
	defer ed.mu.RUnlock()

	kv, exists := ed.data[key]
	if !exists {
		return "", false
	}

	// Check if expired
	if kv.ExpiresAt != nil && time.Now().After(*kv.ExpiresAt) {
		// Key has expired, delete it
		go ed.Delete(key)
		return "", false
	}

	return kv.Value, true
}

// Delete removes a key from the dictionary
func (ed *ExpiringDict) Delete(key string) {
	ed.mu.Lock()
	defer ed.mu.Unlock()
	delete(ed.data, key)
}

// Global dictionary instance
var Dictionary = NewExpiringDict()
