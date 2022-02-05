package server

import (
	"context"
	"sort"
	"sync"

	"github.com/xvello/letsblockit/src/db"
	"github.com/xvello/letsblockit/src/filters"
)

type filterSortOrder uint8

const (
	sortByNameAsc = filterSortOrder(iota)
	sortByUsageDesc
	sortByUpdateDesc
	sortByCreationDesc
)

type filterInfo struct {
	sync.RWMutex
	age     map[string]db.FilterTemplate
	filters []*filters.Filter
	sorted  map[filterSortOrder][]*filters.Filter
	usage   map[string]int64
}

func (s *Server) buildFilterInfo() (*filterInfo, error) {
	info := &filterInfo{
		age:     make(map[string]db.FilterTemplate, len(s.filters.GetFilters())),
		filters: s.filters.GetFilters(),
		sorted:  make(map[filterSortOrder][]*filters.Filter),
		usage:   make(map[string]int64, len(s.filters.GetFilters())),
	}

	return info, s.store.RunCtxTx(context.Background(), func(ctx context.Context, q db.Querier) error {
		// Get templates already registered in the DB
		registeredList, e := q.GetFilterTemplates(ctx)
		if e != nil {
			return e
		}
		for _, t := range registeredList {
			info.age[t.FilterName] = t
		}

		// Check whether new / updated templates need registration
		for _, f := range info.filters {
			hash := f.TemplateHash()
			existing, found := info.age[f.Name]
			if !found {
				if info.age[f.Name], e = q.RegisterNewTemplate(ctx, db.RegisterNewTemplateParams{
					FilterName:   f.Name,
					TemplateHash: hash,
				}); e != nil {
					return e
				}
				_ = s.statsd.Incr("letsblockit.template_changes", []string{"type:new"}, 1)
			} else if existing.TemplateHash != hash {
				if info.age[f.Name], e = q.RegisterUpdatedTemplate(ctx, db.RegisterUpdatedTemplateParams{
					ID:           existing.ID,
					TemplateHash: hash,
				}); e != nil {
					return e
				}
				_ = s.statsd.Incr("letsblockit.template_changes", []string{"type:updated"}, 1)
			}
		}

		// Collect usage stats
		stats, e := q.GetInstanceStats(ctx)
		if e != nil {
			return e
		}
		for _, s := range stats {
			info.usage[s.FilterName] = s.Count
		}

		return nil
	})
}

func (f *filterInfo) getSortedFilters(order filterSortOrder) []*filters.Filter {
	// No sort
	if order == sortByNameAsc {
		return f.filters
	}

	// Cache hit
	f.RLock()
	list, ok := f.sorted[order]
	f.RUnlock()
	if ok {
		return list
	}

	// Cache miss
	f.Lock()
	defer f.Unlock()
	list = append(list, f.filters...)

	switch order {
	case sortByUsageDesc:
		sort.Slice(list, func(i, j int) bool {
			return f.usage[list[i].Name] > f.usage[list[j].Name]
		})
	case sortByUpdateDesc:
		sort.Slice(list, func(i, j int) bool {
			return f.age[list[i].Name].UpdatedAt.Before(f.age[list[j].Name].UpdatedAt)
		})
	case sortByCreationDesc:
		sort.Slice(list, func(i, j int) bool {
			return f.age[list[i].Name].CreatedAt.Before(f.age[list[j].Name].CreatedAt)
		})
	}
	f.sorted[order] = list
	return list
}
