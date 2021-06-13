package cron

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDebug(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/cron/debug", nil)
	w := httptest.NewRecorder()
	debug(w, r)
}
