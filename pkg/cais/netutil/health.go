package netutil

// HealthPayload builds a JSON-serializable /health response with LAN URLs for
// mobile testing. LAN URLs expose the host's RFC1918 topology, so they are only
// included outside production (#131).
func HealthPayload(status, port, env string) map[string]any {
	payload := map[string]any{"status": status}
	if env != "production" {
		payload["lan_urls"] = LANURLs(port)
	}
	return payload
}
