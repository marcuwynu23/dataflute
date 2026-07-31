package scalar

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValue(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want any
	}{
		{"empty", "", ""},
		{"whitespace", "  hello  ", "hello"},
		{"true", "true", true},
		{"true upper", "TRUE", true},
		{"false", "false", false},
		{"null", "null", nil},
		{"tilde", "~", nil},
		{"integer", "42", json.Number("42")},
		{"negative", "-3", json.Number("-3")},
		{"float", "3.14", json.Number("3.14")},
		{"exponent", "1e3", json.Number("1e3")},
		{"text", "hello", "hello"},
		{"spaced number", " 99 ", json.Number("99")},
		{"numeric string", "007", "007"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Value(tc.in)
			if tc.want == nil {
				assert.Nil(t, got)
				return
			}
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestNumber(t *testing.T) {
	require.True(t, ok(number("123")))
	require.True(t, ok(number("-1.5e10")))
	require.False(t, ok(number("")))
	require.False(t, ok(number("abc")))
	require.False(t, ok(number("01")))
}

func ok(_ json.Number, ok bool) bool { return ok }
