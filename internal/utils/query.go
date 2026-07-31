package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

// reservedParams are query parameters handled by the API itself rather
// than treated as field filters.
var reservedParams = map[string]struct{}{
	"q":      {},
	"sort":   {},
	"fields": {},
	"page":   {},
	"limit":  {},
}

// ApplyQuery applies the query features (filter, search, sort, field
// selection and pagination) to the root value of a dataset.
//
// Query features apply when the root value is an array; other root types
// are returned unchanged. The returned value never aliases the input's
// backing array, so the cached dataset is never mutated.
func ApplyQuery(data any, q url.Values) (any, error) {
	if len(q) == 0 {
		return data, nil
	}

	items, ok := data.([]any)
	if !ok {
		return data, nil
	}
	// Copy so sorting below never mutates the cached dataset.
	result := append([]any(nil), items...)

	for key, values := range q {
		if _, reserved := reservedParams[key]; reserved {
			continue
		}
		if len(values) == 0 || values[0] == "" {
			continue
		}
		result = filterExact(result, key, values[0])
	}

	if term := q.Get("q"); term != "" {
		result = filterSearch(result, term)
	}

	if spec := q.Get("sort"); spec != "" {
		if err := sortItems(result, spec); err != nil {
			return nil, err
		}
	}

	if spec := q.Get("fields"); spec != "" {
		result = selectFields(result, spec)
	}

	page := 1
	if raw := q.Get("page"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 {
			return nil, errors.New("invalid page")
		}
		page = n
	}

	if raw := q.Get("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 {
			return nil, errors.New("invalid limit")
		}
		start := (page - 1) * n
		if start >= len(result) {
			result = []any{}
		} else {
			end := start + n
			if end > len(result) {
				end = len(result)
			}
			result = result[start:end]
		}
	}

	return result, nil
}

// filterExact keeps elements whose top-level key exactly matches value.
func filterExact(items []any, key, value string) []any {
	out := make([]any, 0, len(items))
	for _, item := range items {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		v, ok := obj[key]
		if !ok {
			continue
		}
		if scalarEqual(v, value) {
			out = append(out, item)
		}
	}
	return out
}

// scalarEqual reports whether a parsed scalar equals a query string.
func scalarEqual(v any, expected string) bool {
	switch t := v.(type) {
	case string:
		return t == expected
	case json.Number:
		return t.String() == expected
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64) == expected
	case bool:
		return strconv.FormatBool(t) == expected
	case nil:
		return expected == "null" || expected == ""
	default:
		return fmt.Sprint(v) == expected
	}
}

// filterSearch keeps elements whose string fields contain term
// (case-insensitive substring match).
func filterSearch(items []any, term string) []any {
	q := strings.ToLower(term)
	out := make([]any, 0, len(items))
	for _, item := range items {
		if containsTerm(item, q) {
			out = append(out, item)
		}
	}
	return out
}

func containsTerm(v any, q string) bool {
	switch t := v.(type) {
	case map[string]any:
		for _, fv := range t {
			if containsTerm(fv, q) {
				return true
			}
		}
		return false
	case []any:
		for _, ev := range t {
			if containsTerm(ev, q) {
				return true
			}
		}
		return false
	case string:
		return strings.Contains(strings.ToLower(t), q)
	case json.Number:
		return strings.Contains(t.String(), q)
	default:
		return false
	}
}

type sortKey struct {
	name string
	desc bool
}

// sortItems stably sorts items by one or more top-level keys.
// A leading "-" on a key sorts that key in descending order.
func sortItems(items []any, spec string) error {
	var keys []sortKey
	for _, part := range strings.Split(spec, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		desc := strings.HasPrefix(part, "-")
		key := strings.TrimPrefix(part, "-")
		if key == "" {
			return errors.New("invalid sort key")
		}
		keys = append(keys, sortKey{name: key, desc: desc})
	}
	if len(keys) == 0 {
		return nil
	}

	sort.SliceStable(items, func(i, j int) bool {
		for _, k := range keys {
			c := compareByKey(items[i], items[j], k.name)
			if c == 0 {
				continue
			}
			if k.desc {
				return c > 0
			}
			return c < 0
		}
		return false
	})
	return nil
}

func compareByKey(a, b any, key string) int {
	oa, okA := a.(map[string]any)
	ob, okB := b.(map[string]any)
	if !okA || !okB {
		return 0
	}
	va, hasA := oa[key]
	vb, hasB := ob[key]
	switch {
	case !hasA && !hasB:
		return 0
	case !hasA:
		return -1
	case !hasB:
		return 1
	}
	return compareValues(va, vb)
}

// compareValues compares two parsed scalar values.
func compareValues(a, b any) int {
	if an, ok := a.(json.Number); ok {
		if bn, ok := b.(json.Number); ok {
			af, errA := an.Float64()
			bf, errB := bn.Float64()
			if errA == nil && errB == nil {
				switch {
				case af < bf:
					return -1
				case af > bf:
					return 1
				default:
					return 0
				}
			}
		}
	}
	as, okA := a.(string)
	bs, okB := b.(string)
	if okA && okB {
		switch {
		case as < bs:
			return -1
		case as > bs:
			return 1
		default:
			return 0
		}
	}
	return strings.Compare(fmt.Sprint(a), fmt.Sprint(b))
}

// selectFields keeps only the requested top-level fields of each element.
func selectFields(items []any, spec string) []any {
	var fields []string
	for _, f := range strings.Split(spec, ",") {
		if f = strings.TrimSpace(f); f != "" {
			fields = append(fields, f)
		}
	}

	out := make([]any, 0, len(items))
	for _, item := range items {
		obj, ok := item.(map[string]any)
		if !ok {
			out = append(out, item)
			continue
		}
		sel := make(map[string]any, len(fields))
		for _, f := range fields {
			if v, ok := obj[f]; ok {
				sel[f] = v
			}
		}
		out = append(out, sel)
	}
	return out
}
