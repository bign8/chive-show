package system

import "net/http"

// System interface for underlying system implementations
type System interface {

	// Store return persistent storage API (should handle memcache if available)
	Store() Store

	// Fetch return client fetchers
	Fetch() *http.Client

	// Defer execute a function in deferred context
	Defer() error
}

// Store deals with system persistant storage
type Store interface {
	Create() error
	Get() error
	Update() error
	Delete() error
	Query() error
}

// TODO: make store much more robust!!!
