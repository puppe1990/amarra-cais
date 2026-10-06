package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// #320: upgrade sobe go.mod mas o amarra.js vendored só atualiza com
// `amarra-cais pwa` — o doctor precisa comparar as duas versões.
func checkBundleVersion(dir string) doctorCheck {
	const name = "amarra.js bundle version"
	bundlePath := filepath.Join(dir, "web/static/js/amarra.js")
	head, err := os.ReadFile(bundlePath)
	if err != nil {
		return doctorCheck{Name: name, OK: true, Detail: "skipped (no web/static/js/amarra.js)"}
	}
	marker := bundleVersionMarker.FindStringSubmatch(string(head[:min(len(head), 512)]))
	if marker == nil {
		return doctorCheck{
			Name: name, Optional: true,
			Detail:  "web/static/js/amarra.js sem marcador de versão",
			FixHint: "amarra-cais pwa",
		}
	}
	want := appCaisVersion(dir)
	if want == "" {
		return doctorCheck{Name: name, OK: true, Detail: "skipped (no framework version in go.mod)"}
	}
	if marker[1] != want {
		return doctorCheck{
			Name: name, Optional: true,
			Detail:  fmt.Sprintf("bundle v%s defasado do go.mod v%s", marker[1], want),
			FixHint: "amarra-cais pwa",
		}
	}
	return doctorCheck{Name: name, OK: true}
}

var bundleVersionMarker = regexp.MustCompile(`/\* amarra-cais v(\d+\.\d+\.\d+) \*/`)

var gomodCaisRequire = regexp.MustCompile(`github.com/puppe1990/amarra-cais v?(\d+\.\d+\.\d+)`)

func appCaisVersion(dir string) string {
	data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return ""
	}
	m := gomodCaisRequire.FindStringSubmatch(string(data))
	if m == nil {
		return ""
	}
	return m[1]
}
