package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestCLI_UsageErrors_UseAmarraCais(t *testing.T) {
	c := &CLI{Out: &bytes.Buffer{}}
	cases := [][]string{
		{"db"},
		{"jobs"},
		{"jobs", "retry"},
		{"destroy"},
	}
	for _, args := range cases {
		err := c.Run(args)
		if err == nil {
			t.Fatalf("%v: expected usage error", args)
		}
		msg := err.Error()
		if !strings.Contains(msg, "usage: amarra-cais") {
			t.Errorf("%v: missing %q in %q", args, "usage: amarra-cais", msg)
		}
		if strings.Contains(msg, "usage: cais ") {
			t.Errorf("%v: still documents %q in %q", args, "usage: cais ", msg)
		}
	}
}
