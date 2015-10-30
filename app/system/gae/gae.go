package gae

import (
	"net/http"

	"github.com/bign8/chive-show/app/system"

	// "appengine"
)

// New create new gae storage engine
func New(c interface{}) system.System { // *appengine.Context
	// TODO: typecase c to appengine.Context
	return &gaeSystem{}
}

type gaeSystem struct {
	c error // *appengine.Context
}

func (gae *gaeSystem) Store() system.Store {
	return nil
}

func (gae *gaeSystem) Fetch() *http.Client {
	return nil
}
