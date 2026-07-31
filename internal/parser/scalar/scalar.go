// Package scalar converts raw string values into typed Go values.
package scalar

import (
	"encoding/json"
	"strings"
)

// Value converts a raw string into the most specific scalar type it
// represents: booleans become bool, numbers become json.Number, the
// literals "null" and "~" become nil, and everything else stays a string.
func Value(s string) any {
	s = strings.TrimSpace(s)
	switch strings.ToLower(s) {
	case "true":
		return true
	case "false":
		return false
	case "null", "~":
		return nil
	}
	if n, ok := number(s); ok {
		return n
	}
	return s
}

// number reports whether s is a valid JSON number literal.
func number(s string) (json.Number, bool) {
	if s == "" || !json.Valid([]byte(s)) {
		return "", false
	}
	var n json.Number
	if err := json.Unmarshal([]byte(s), &n); err != nil {
		return "", false
	}
	return n, true
}
