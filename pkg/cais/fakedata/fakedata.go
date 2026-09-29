// Package fakedata generates realistic sample data for demo seeds, forms and
// fixtures. Values differ on every call unless Seed pins a reproducible
// sequence.
package fakedata

import (
	"sync/atomic"

	"github.com/brianvoe/gofakeit/v7"
)

// dateLayout matches <input type="date"> values and SQLite date text.
const dateLayout = "2006-01-02"

// Seed replaces the faker, so it lives behind an atomic pointer: one
// goroutine may call Seed while handler goroutines keep generating.
// gofakeit.New returns a locked Faker, safe for concurrent use.
var faker atomic.Pointer[gofakeit.Faker]

func init() { faker.Store(gofakeit.New(0)) }

// Seed pins the generators to a reproducible sequence, useful in tests and
// demo data. Seed(0) goes back to random values.
func Seed(n int64) { faker.Store(gofakeit.New(uint64(n))) }

// Name returns a first and last name, e.g. "Ada Lovelace".
func Name() string { return faker.Load().Name() }

// Title returns a human-readable title, e.g. "Crime and Punishment".
func Title() string { return faker.Load().BookTitle() }

// Email returns an email address.
func Email() string { return faker.Load().Email() }

// Sentence returns a capitalized sentence ending in punctuation.
func Sentence() string { return faker.Load().Sentence() }

// URL returns an absolute URL with an http(s) scheme.
func URL() string { return faker.Load().URL() }

// Date returns a past-or-current date formatted as YYYY-MM-DD.
func Date() string { return faker.Load().Date().Format(dateLayout) }
