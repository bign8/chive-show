package system

import (
	"errors"
	"net/http"
)

// System interface for underlying system implementations
type System interface {

	// Store return persistent storage API (should handle memcache if available)
	Store() Store

	// Fetch return client fetchers
	Fetch() *http.Client

	// Defer execute a function in deferred context
	// Defer() error
}

// ErrDoesNotExist if item key does not exist
var ErrDoesNotExist = errors.New("Item does not exist")

// Store deals with system persistant storage
type Store interface {
	// Get returns item to dst if found
	Get(*Key, interface{}) error

	// GetMulti populates dst if keys found
	GetMulti([]*Key, []interface{}) error

	// Put stores item
	Put(*Key, interface{}) error

	// PutMulti stores items
	PutMulti([]*Key, []interface{}) error

	// Delete removes item
	Delete(*Key) error

	// DeleteMulti removes multiple items
	DeleteMulti([]*Key) error

	// Query TODO: yeah...
	Query() error
}

// Key is type of an item key
type Key string
