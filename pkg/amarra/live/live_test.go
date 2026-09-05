package live

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_returns501(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/amarra/live", nil)
	Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusNotImplemented {
		t.Fatalf("code %d", rr.Code)
	}
}
