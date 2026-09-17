package pagination

import "math"

// Meta holds offset pagination state for list pages.
type Meta struct {
	Page     int
	PerPage  int
	Total    int
	HasPrev  bool
	HasNext  bool
	PrevPage int
	NextPage int
}

const defaultPerPage = 25

// New builds pagination metadata for a list page. A page beyond the last one
// (hand-crafted ?page=) is clamped after the COUNT so HasNext/NextPage stay
// correct and offsets cannot overflow (#135).
func New(page, perPage, total int) Meta {
	if perPage < 1 {
		perPage = defaultPerPage
	}
	if total < 0 {
		total = 0
	}
	lastPage := total / perPage
	if total%perPage != 0 {
		lastPage++
	}
	if lastPage < 1 {
		lastPage = 1
	}
	if page < 1 {
		page = 1
	}
	if page > lastPage {
		page = lastPage
	}
	return Meta{
		Page:     page,
		PerPage:  perPage,
		Total:    total,
		HasPrev:  page > 1,
		HasNext:  page < lastPage,
		PrevPage: page - 1,
		NextPage: page + 1,
	}
}

// Offset returns the SQL OFFSET for page and perPage. Values that would
// overflow int saturate at the largest aligned offset instead (#135).
func Offset(page, perPage int) int {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = defaultPerPage
	}
	if page-1 > math.MaxInt/perPage {
		return (math.MaxInt / perPage) * perPage
	}
	return (page - 1) * perPage
}
