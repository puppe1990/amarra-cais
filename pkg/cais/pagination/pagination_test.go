package pagination

import (
	"math"
	"testing"
)

func TestNew_normalizesPageAndPerPage(t *testing.T) {
	m := New(0, 0, 100)
	if m.Page != 1 {
		t.Errorf("Page = %d, want 1", m.Page)
	}
	if m.PerPage != 25 {
		t.Errorf("PerPage = %d, want 25", m.PerPage)
	}
	if !m.HasNext {
		t.Error("expected HasNext")
	}
	if m.HasPrev {
		t.Error("unexpected HasPrev on first page")
	}
}

func TestNew_lastPage(t *testing.T) {
	m := New(4, 25, 100)
	if m.HasNext {
		t.Error("unexpected HasNext on last page")
	}
	if !m.HasPrev {
		t.Error("expected HasPrev")
	}
	if m.PrevPage != 3 || m.NextPage != 5 {
		t.Errorf("PrevPage=%d NextPage=%d", m.PrevPage, m.NextPage)
	}
}

func TestOffset(t *testing.T) {
	if got := Offset(3, 25); got != 50 {
		t.Errorf("Offset = %d, want 50", got)
	}
	if got := Offset(0, 0); got != 0 {
		t.Errorf("Offset(0,0) = %d, want 0", got)
	}
}

// #135: a hand-crafted ?page= could make (page-1)*perPage overflow int,
// producing a negative/absurd OFFSET and a wrong HasNext.
func TestOffset_saturatesInsteadOfOverflowing(t *testing.T) {
	want := (math.MaxInt / 25) * 25
	if got := Offset(1<<62, 25); got != want {
		t.Fatalf("Offset(1<<62, 25) = %d, want saturated %d", got, want)
	}
	if small := Offset(3, 25); small != 50 {
		t.Fatalf("Offset(3, 25) = %d, want 50", small)
	}
}

func TestNew_clampsHugePageToLastPage(t *testing.T) {
	pg := New(1<<62, 25, 120) // 5 pages
	if pg.Page != 5 {
		t.Fatalf("Page = %d, want clamped to 5", pg.Page)
	}
	if pg.HasNext {
		t.Error("HasNext should be false on the last page")
	}
	if pg.HasPrev != true || pg.PrevPage != 4 {
		t.Fatalf("HasPrev/PrevPage = %v/%d, want true/4", pg.HasPrev, pg.PrevPage)
	}

	empty := New(99, 25, 0)
	if empty.Page != 1 || empty.HasNext || empty.HasPrev {
		t.Fatalf("empty list = %+v, want page 1 without prev/next", empty)
	}
}
