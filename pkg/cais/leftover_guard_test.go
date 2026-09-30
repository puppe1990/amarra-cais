package cais

import (
	"os"
	"testing"
)

// #265: barcode (Open Food Facts) and money.FormatBRL belonged to one app,
// not the generic framework. They live in the app if someone still needs them.
func TestPkgCais_omitsBarcodeAndMoney(t *testing.T) {
	for _, dir := range []string{"barcode", "money"} {
		if _, err := os.Stat(dir); err == nil {
			t.Errorf("pkg/cais/%s still exists (#265)", dir)
		}
	}
}
