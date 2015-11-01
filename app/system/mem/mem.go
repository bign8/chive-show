package mem

import (
	"log"

	"github.com/bign8/chive-show/app/system"
)

// New create new gae storage engine
func New(_ interface{}) system.System {
	log.Printf("Creating new System store")
	store := memStore{
		data: make(map[system.Key]interface{}),
	}
	return &memSystem{
		store: store,
	}
}
