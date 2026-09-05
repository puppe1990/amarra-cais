package amarra

import "net/http"

const HeaderDrive = "Amarra-Drive" // value "true"

func IsDrive(r *http.Request) bool {
	return r.Header.Get(HeaderDrive) == "true"
}
