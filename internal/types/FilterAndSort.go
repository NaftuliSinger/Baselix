package types

import (
	"fmt"
	"sort"
	"strings"
)

type Filter struct {
	Field    string
	Operator string
	Value    string
}

type Sort struct {
	Field string
	Dir   string
}

func ParseFilters(filterStr string) ([]Filter, error) {
	var filters []Filter

	if filterStr == "" {
		return nil, nil
	}

	parts := strings.Split(filterStr, ",")

	for _, p := range parts {
		f := strings.SplitN(p, ":", 3)
		if len(f) != 3 {
			return nil, fmt.Errorf("invalid filter format: %q", p)
		}

		filters = append(filters, Filter{
			Field:    f[0],
			Operator: f[1],
			Value:    f[2],
		})
	}

	return filters, nil
}

func ParseSorts(sortStr string) ([]Sort, error) {
	var sorts []Sort

	if sortStr == "" {
		return nil, nil
	}

	parts := strings.Split(sortStr, ",")

	for _, p := range parts {
		s := strings.SplitN(p, ":", 2)
		if len(s) != 2 {
			return nil, fmt.Errorf("invalid sort format: %q", p)
		}
		// if dir is not "asc" or "desc", default to "asc"
		if !strings.EqualFold(s[1], "asc") && !strings.EqualFold(s[1], "desc") {
			s[1] = "asc"
		}

		sorts = append(sorts, Sort{
			Field: s[0],
			Dir:   s[1],
		})
	}

	return sorts, nil
}

func ApplySorts(records []RecordResponse, sorts []Sort) ([]RecordResponse, error) {
	// Validate that every sort field exists in at least one record.
	if len(records) > 0 {
		for _, s := range sorts {
			found := false
			for _, r := range records {
				for _, f := range r.Values {
					if f.Key == s.Field {
						found = true
						break
					}
				}
				if found {
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("sort field %q does not exist", s.Field)
			}
		}
	}

	// Loop over sorts in reverse order to apply them correctly.
	for i := len(sorts) - 1; i >= 0; i-- {
		s := sorts[i]
		sort.SliceStable(records, func(a, b int) bool {
			var valA, valB any
			for _, f := range records[a].Values {
				if f.Key == s.Field {
					valA = f.Value
					break
				}
			}
			for _, f := range records[b].Values {
				if f.Key == s.Field {
					valB = f.Value
					break
				}
			}
			less := compareValues(valA, valB)
			if strings.EqualFold(s.Dir, "desc") {
				return !less
			}
			return less
		})
	}
	return records, nil
}

func compareValues(a, b any) bool {
	aFloat, aOk := toFloat64(a)
	bFloat, bOk := toFloat64(b)
	if aOk && bOk {
		return aFloat < bFloat
	}
	return fmt.Sprintf("%v", a) < fmt.Sprintf("%v", b)
}

func toFloat64(v any) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case float32:
		return float64(n), true
	case float64:
		return n, true
	}
	return 0, false
}
