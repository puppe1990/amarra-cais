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

func TestWriteOp_appendIncludesTarget(t *testing.T) {
	rr := httptest.NewRecorder()
	if err := WriteOp(rr, Op{Kind: "append", Target: "chat-history", HTML: "<p>hi</p>"}); err != nil {
		t.Fatal(err)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "id: chat-history") {
		t.Fatalf("missing target id, got %q", body)
	}
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

func TestWriteOp_beforeAndAfter(t *testing.T) {
	rr := httptest.NewRecorder()
	if err := WriteOp(rr, Op{Kind: "before", Target: "item-1", HTML: "<li>n</li>"}); err != nil {
		t.Fatal(err)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "event: before") || !strings.Contains(body, "id: item-1") {
		t.Fatalf("%q", body)
	}
}

func TestWriteHTTP_setsStreamContentType(t *testing.T) {
	rr := httptest.NewRecorder()
	if err := WriteHTTP(rr, Op{Kind: "append", Target: "list", HTML: "<li>x</li>"}); err != nil {
		t.Fatal(err)
	}
	ct := rr.Header().Get("Content-Type")
	if !strings.Contains(ct, "vnd.amarra-stream") {
		t.Fatalf("content-type %q", ct)
	}
	if !strings.Contains(rr.Body.String(), "event: append") {
		t.Fatalf("%q", rr.Body.String())
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
