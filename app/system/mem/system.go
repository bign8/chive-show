package mem

import (
	"net/http"

	"github.com/bign8/chive-show/app/system"
)

type memSystem struct {
	store memStore
}

func (mem *memSystem) Store() system.Store {
	return &mem.store
}

func (mem *memSystem) Fetch() *http.Client {
	return nil
}

func (mem *memSystem) Defer() error {
	return nil
}
