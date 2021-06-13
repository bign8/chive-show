package cron

import (
	"net/http"
	"net/http/httputil"

	"github.com/bign8/chive-show/models"
)

func MineHandler(store models.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// env := os.Environ()
		// str := strings.Join(env, "\n")
		// http.Error(w, str, http.StatusOK)

		bits, err := httputil.DumpRequest(r, true)
		if err != nil {
			panic(err)
		}
		w.Write(bits)
	}
}
