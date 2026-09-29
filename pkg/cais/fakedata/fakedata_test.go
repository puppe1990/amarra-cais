package fakedata

import (
	"strings"
	"sync"
	"testing"
	"time"
	"unicode"
)

// sample captures one value from every generator at a single point in time.
type sample struct {
	name     string
	title    string
	email    string
	sentence string
	url      string
	date     string
}

func sampleAll() sample {
	return sample{
		name:     Name(),
		title:    Title(),
		email:    Email(),
		sentence: Sentence(),
		url:      URL(),
		date:     Date(),
	}
}

// Seed is process-wide, so these tests stay serial (no t.Parallel).

func TestSeed_sameSeedSameValues(t *testing.T) {
	Seed(42)
	first := sampleAll()
	Seed(42)
	second := sampleAll()
	if first != second {
		t.Errorf("Seed(42) is not reproducible:\nfirst:  %+v\nsecond: %+v", first, second)
	}
}

func TestSeed_differentSeedsDiffer(t *testing.T) {
	Seed(1)
	first := sampleAll()
	Seed(2)
	second := sampleAll()
	if first == second {
		t.Errorf("Seed(1) and Seed(2) produced identical samples: %+v", first)
	}
}

func TestName_isFirstAndLastName(t *testing.T) {
	Seed(1)
	name := Name()
	if parts := strings.Fields(name); len(parts) < 2 {
		t.Errorf("Name() = %q, want at least two words", name)
	}
}

func TestTitle_nonEmpty(t *testing.T) {
	Seed(1)
	if title := Title(); strings.TrimSpace(title) == "" {
		t.Error("Title() is empty")
	}
}

func TestEmail_containsAtSign(t *testing.T) {
	Seed(1)
	if email := Email(); !strings.Contains(email, "@") {
		t.Errorf("Email() = %q, want an @ sign", email)
	}
}

func TestSentence_isCapitalized(t *testing.T) {
	Seed(1)
	sentence := Sentence()
	if sentence == "" {
		t.Fatal("Sentence() is empty")
	}
	if first := rune(sentence[0]); !unicode.IsUpper(first) {
		t.Errorf("Sentence() = %q, want a capitalized first letter", sentence)
	}
}

func TestURL_hasScheme(t *testing.T) {
	Seed(1)
	url := URL()
	if !strings.HasPrefix(url, "http") || !strings.Contains(url, "://") {
		t.Errorf("URL() = %q, want an http(s) scheme", url)
	}
}

func TestDate_matchesLayout(t *testing.T) {
	Seed(1)
	date := Date()
	if _, err := time.Parse("2006-01-02", date); err != nil {
		t.Errorf("Date() = %q, want YYYY-MM-DD: %v", date, err)
	}
}

func TestGenerators_areConcurrencySafe(t *testing.T) {
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 200 {
				_ = sampleAll()
			}
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := range 200 {
			Seed(int64(i))
		}
	}()
	wg.Wait()
}
