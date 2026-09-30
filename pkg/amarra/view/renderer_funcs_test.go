package view

import "testing"

// #265: formatMoney/FormatBRL was a product leftover registered on every
// renderer. Apps that need a currency helper register it locally.
func TestTemplateFuncs_omitFormatMoney(t *testing.T) {
	funcs := templateFuncs(nil)
	if _, ok := funcs["formatMoney"]; ok {
		t.Error("formatMoney must not be a default view helper (#265)")
	}
}
