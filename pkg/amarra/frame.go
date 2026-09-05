package amarra

import "net/http"

const HeaderFrame = "Amarra-Frame" // frame id

func FrameID(r *http.Request) string {
	return r.Header.Get(HeaderFrame)
}
