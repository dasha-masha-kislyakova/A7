package audit

import (
	"sort"
	"strings"
	"sync"
	"time"
)

type Repo interface {
	Insert(e Event) (Event, error)
	List(q Query) ([]Event, int) // items, total
}

type memRepo struct {
	mu    sync.RWMutex
	items []Event
	next  int64
}

func NewMemRepo() Repo {
	return &memRepo{items: make([]Event, 0, 1024), next: 1}
}

func (r *memRepo) Insert(e Event) (Event, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	e.ID = r.next
	if e.At.IsZero() {
		e.At = time.Now()
	}
	r.next++
	r.items = append(r.items, e)
	return e, nil
}

func (r *memRepo) List(q Query) ([]Event, int) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	flt := make([]Event, 0, len(r.items))
	for _, it := range r.items {
		if q.SourceService != "" && !strings.EqualFold(it.SourceService, q.SourceService) {
			continue
		}
		if q.TargetService != "" && !strings.EqualFold(it.TargetService, q.TargetService) {
			continue
		}
		if q.URI != "" && !strings.Contains(strings.ToLower(it.URI), strings.ToLower(q.URI)) {
			continue
		}
		if q.HTTPStatus != nil && it.HTTPStatus != *q.HTTPStatus {
			continue
		}
		if q.DateFrom != nil && it.At.Before(*q.DateFrom) {
			continue
		}
		if q.DateTo != nil && it.At.After(*q.DateTo) {
			continue
		}
		if q.MinDurationMs != nil && it.DurationMs < *q.MinDurationMs {
			continue
		}
		if q.UserID != "" && !strings.EqualFold(it.UserID, q.UserID) {
			continue
		}
		flt = append(flt, it)
	}

	switch q.SortBy {
	case "source":
		sort.Slice(flt, func(i, j int) bool {
			if q.SortOrder == "desc" {
				return flt[i].SourceService > flt[j].SourceService
			}
			return flt[i].SourceService < flt[j].SourceService
		})
	case "target":
		sort.Slice(flt, func(i, j int) bool {
			if q.SortOrder == "desc" {
				return flt[i].TargetService > flt[j].TargetService
			}
			return flt[i].TargetService < flt[j].TargetService
		})
	case "status":
		sort.Slice(flt, func(i, j int) bool {
			if q.SortOrder == "desc" {
				return flt[i].HTTPStatus > flt[j].HTTPStatus
			}
			return flt[i].HTTPStatus < flt[j].HTTPStatus
		})
	case "duration":
		sort.Slice(flt, func(i, j int) bool {
			if q.SortOrder == "desc" {
				return flt[i].DurationMs > flt[j].DurationMs
			}
			return flt[i].DurationMs < flt[j].DurationMs
		})
	case "user":
		sort.Slice(flt, func(i, j int) bool {
			if q.SortOrder == "desc" {
				return flt[i].UserID > flt[j].UserID
			}
			return flt[i].UserID < flt[j].UserID
		})
	}

	total := len(flt)
	page := q.Page
	if page < 1 {
		page = 1
	}
	size := q.PageSize
	if size != 10 && size != 50 && size != 100 {
		size = 10
	}
	start := (page - 1) * size
	if start > total {
		return []Event{}, total
	}
	end := start + size
	if end > total {
		end = total
	}
	return flt[start:end], total
}
