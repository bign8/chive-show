package cron

import (
	"net/http"

	"github.com/bign8/chive-show/models"
)

var latestVersion = 0

func MigrateHandler(store models.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "todo", http.StatusNotImplemented)
	}
}
