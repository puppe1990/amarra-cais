package devlog

import (
	"fmt"
	"html"
	"net/http"

	"github.com/puppe1990/amarra-cais/pkg/cais"
)

func Register(r *cais.Router, env string, buf *Buffer) {
	if !Enabled(env) || buf == nil {
		r.Get("/logs", func(w http.ResponseWriter, req *http.Request) {
			http.NotFound(w, req)
		})
		return
	}
	h := LocalOnly(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		serveLogs(w, req, buf)
	}))
	r.Get("/logs", h.ServeHTTP)
}

func serveLogs(w http.ResponseWriter, r *http.Request, buf *Buffer) {
	body := html.EscapeString(FormatForDisplay(buf.Text()))
	if body == "" {
		body = "(no logs yet)"
	}

	if r.URL.Query().Get("partial") == "1" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = fmt.Fprint(w, body)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = fmt.Fprintf(w, pageHTML, body)
}

const pageHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Cais Logs</title>
  <style>
    body { margin: 0; font-family: ui-monospace, SFMono-Regular, Menlo, monospace; background: #081014; color: #f3ead8; }
    header { padding: 1rem 1.25rem; border-bottom: 1px solid rgba(201,137,58,0.35); display: flex; justify-content: space-between; align-items: center; }
    h1 { margin: 0; font-size: 0.95rem; color: #f3ead8; font-weight: 600; }
    .badge { font-size: 0.7rem; color: #c9893a; text-transform: uppercase; letter-spacing: 0.08em; }
    #logs { margin: 0; padding: 1.25rem; min-height: calc(100vh - 4rem); overflow: auto; white-space: pre-wrap; word-break: break-word; }
  </style>
</head>
<body>
  <header>
    <h1>Cais Logs</h1>
    <span class="badge">development · localhost only · auto-refresh 2s</span>
  </header>
  <pre id="logs">%s</pre>
  <script>
    setInterval(function () {
      fetch("/logs?partial=1").then(function (r) { return r.text(); }).then(function (t) {
        var el = document.getElementById("logs");
        if (el) el.textContent = t;
      }).catch(function () {});
    }, 2000);
  </script>
</body>
</html>
`
