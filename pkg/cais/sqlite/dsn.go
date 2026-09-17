package sqlite

import "strings"

// DSN appends the Cais pragmas to a modernc SQLite DSN.
//
// Configure applies PRAGMAs with Exec on one pooled connection; they do not
// survive a driver reconnect (ErrBadConn), and the app would silently lose
// foreign_keys enforcement and busy_timeout retries (#122). With the pragmas
// in the DSN, every connection the driver opens inherits them.
func DSN(path string) string {
	if path == "" || strings.Contains(path, "_pragma=") {
		return path
	}
	sep := "?"
	if strings.Contains(path, "?") {
		sep = "&"
	}
	return path + sep + "_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)"
}
