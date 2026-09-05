package live

import "net/http"

func Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "amarra live is not enabled in this release", http.StatusNotImplemented)
	})
}
