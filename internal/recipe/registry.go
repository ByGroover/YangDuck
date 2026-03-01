package recipe

import (
	"sort"
	"strings"
)

type Registry struct {
	recipes map[string]Recipe
	byType  map[RecipeType][]Recipe
}

type ListOptions struct {
	Type     RecipeType
	Tag      string
	Page     int
	PageSize int
	SortBy   string // "name", "popularity", "added_at"
}

type ListResult struct {
	Items      []Recipe
	Total      int
	Page       int
	TotalPages int
}

func NewRegistry() *Registry {
	return &Registry{
		recipes: make(map[string]Recipe),
		byType:  make(map[RecipeType][]Recipe),
	}
}

func (r *Registry) Add(recipes ...Recipe) {
	for _, rec := range recipes {
		r.recipes[rec.ID] = rec
		r.byType[rec.Type] = append(r.byType[rec.Type], rec)
	}
}

func (r *Registry) Get(id string) (Recipe, bool) {
	rec, ok := r.recipes[id]
	return rec, ok
}

func (r *Registry) ListByType(t RecipeType) []Recipe {
	return r.byType[t]
}

func (r *Registry) All() []Recipe {
	out := make([]Recipe, 0, len(r.recipes))
	for _, rec := range r.recipes {
		out = append(out, rec)
	}
	return out
}

func (r *Registry) Search(keyword string) []Recipe {
	keyword = strings.ToLower(keyword)
	var results []Recipe
	for _, rec := range r.recipes {
		if matchesKeyword(rec, keyword) {
			results = append(results, rec)
		}
	}
	return results
}

func matchesKeyword(rec Recipe, kw string) bool {
	if strings.Contains(strings.ToLower(rec.Name), kw) {
		return true
	}
	if strings.Contains(strings.ToLower(rec.Description), kw) {
		return true
	}
	if strings.Contains(strings.ToLower(rec.ID), kw) {
		return true
	}
	for _, tag := range rec.Tags {
		if strings.Contains(strings.ToLower(tag), kw) {
			return true
		}
	}
	return false
}

func (r *Registry) Count() int {
	return len(r.recipes)
}

func (r *Registry) Bundles() []Recipe {
	return r.byType[TypeBundle]
}

func (r *Registry) List(opts ListOptions) ListResult {
	var filtered []Recipe
	for _, rec := range r.recipes {
		if opts.Type != "" && rec.Type != opts.Type {
			continue
		}
		if opts.Tag != "" {
			found := false
			for _, t := range rec.Tags {
				if strings.EqualFold(t, opts.Tag) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		filtered = append(filtered, rec)
	}

	switch opts.SortBy {
	case "popularity":
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].Popularity > filtered[j].Popularity
		})
	case "added_at":
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].AddedAt > filtered[j].AddedAt
		})
	default:
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].Name < filtered[j].Name
		})
	}

	total := len(filtered)
	pageSize := opts.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	page := opts.Page
	if page <= 0 {
		page = 1
	}
	totalPages := (total + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}

	start := (page - 1) * pageSize
	if start >= total {
		return ListResult{Total: total, Page: page, TotalPages: totalPages}
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	return ListResult{
		Items:      filtered[start:end],
		Total:      total,
		Page:       page,
		TotalPages: totalPages,
	}
}

func (r *Registry) CountByType() map[RecipeType]int {
	counts := make(map[RecipeType]int)
	for _, rec := range r.recipes {
		counts[rec.Type]++
	}
	return counts
}

func (r *Registry) Featured() []Recipe {
	var out []Recipe
	for _, rec := range r.recipes {
		if rec.Featured {
			out = append(out, rec)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Popularity > out[j].Popularity
	})
	return out
}

func (r *Registry) RecentlyAdded(limit int) []Recipe {
	all := r.All()
	sort.Slice(all, func(i, j int) bool {
		return all[i].AddedAt > all[j].AddedAt
	})
	if limit > 0 && limit < len(all) {
		all = all[:limit]
	}
	return all
}

func (r *Registry) Related(id string, limit int) []Recipe {
	rec, ok := r.recipes[id]
	if !ok || len(rec.Tags) == 0 {
		return nil
	}

	tagSet := make(map[string]bool)
	for _, t := range rec.Tags {
		tagSet[strings.ToLower(t)] = true
	}

	type scored struct {
		recipe Recipe
		score  int
	}
	var candidates []scored
	for _, other := range r.recipes {
		if other.ID == id {
			continue
		}
		s := 0
		for _, t := range other.Tags {
			if tagSet[strings.ToLower(t)] {
				s++
			}
		}
		if s > 0 {
			candidates = append(candidates, scored{other, s})
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].score != candidates[j].score {
			return candidates[i].score > candidates[j].score
		}
		return candidates[i].recipe.Popularity > candidates[j].recipe.Popularity
	})

	var out []Recipe
	for i, c := range candidates {
		if limit > 0 && i >= limit {
			break
		}
		out = append(out, c.recipe)
	}
	return out
}
