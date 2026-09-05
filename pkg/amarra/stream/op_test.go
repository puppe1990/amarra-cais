package stream

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteOp_append(t *testing.T) {
	rr := httptest.NewRecorder()
	if err := WriteOp(rr, Op{Kind: "append", Target: "chat-history", HTML: "<p>hi</p>"}); err != nil {
		t.Fatal(err)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "event: append") || !strings.Contains(body, "data: <p>hi</p>") {
		t.Fatalf("%q", body)
	}
}

func TestWriteOp_toast(t *testing.T) {
	rr := httptest.NewRecorder()
	if err := WriteOp(rr, Op{Kind: "toast", HTML: "Saved!"}); err != nil {
		t.Fatal(err)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "event: toast") || !strings.Contains(body, "data: Saved!") {
		t.Fatalf("%q", body)
	}
}

func TestWriteOp_remove(t *testing.T) {
	rr := httptest.NewRecorder()
	if err := WriteOp(rr, Op{Kind: "remove", Target: "item-1"}); err != nil {
		t.Fatal(err)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "event: remove") || !strings.Contains(body, "data: ") {
		t.Fatalf("%q", body)
	}
}
