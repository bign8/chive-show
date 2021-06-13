package cron

import (
	"net/http"
	"net/http/httputil"

	"github.com/bign8/chive-show/models"
)

func RebuildHandler(store models.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bits, err := httputil.DumpRequest(r, true)
		if err != nil {
			panic(err)
		}
		w.Write(bits)
		// http.Error(w, "TODO", http.StatusNotImplemented)
	}
}
