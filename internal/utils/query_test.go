package utils

import (
	"encoding/json"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func decode(t *testing.T, raw string) any {
	t.Helper()
	dec := json.NewDecoder(strings.NewReader(raw))
	dec.UseNumber()
	var v any
	require.NoError(t, dec.Decode(&v))
	return v
}

func values(qs string) url.Values {
	v, _ := url.ParseQuery(qs)
	return v
}

func users(t *testing.T) []any {
	raw := `[
		{"id":1,"name":"John","country":"PH","score":10},
		{"id":2,"name":"Jane","country":"US","score":20},
		{"id":3,"name":"John","country":"US","score":30},
		{"id":4,"name":"Zoe","country":"PH","score":5}
	]`
	return decode(t, raw).([]any)
}

func TestApplyQuery_NoQuery(t *testing.T) {
	data := users(t)
	out, err := ApplyQuery(data, nil)
	require.NoError(t, err)
	assert.Len(t, out, 4)
}

func TestApplyQuery_DoesNotMutateCachedData(t *testing.T) {
	data := users(t)
	_, err := ApplyQuery(data, values("sort=-id"))
	require.NoError(t, err)

	// Original order must be unchanged.
	assert.Equal(t, json.Number("1"), data[0].(map[string]any)["id"])
}

func TestApplyQuery_Filter(t *testing.T) {
	out, err := ApplyQuery(users(t), values("country=PH"))
	require.NoError(t, err)
	items := out.([]any)
	assert.Len(t, items, 2)

	out, err = ApplyQuery(users(t), values("country=PH&name=John"))
	require.NoError(t, err)
	assert.Len(t, out.([]any), 1)
}

func TestApplyQuery_Search(t *testing.T) {
	out, err := ApplyQuery(users(t), values("q=zoe"))
	require.NoError(t, err)
	assert.Len(t, out.([]any), 1)

	out, err = ApplyQuery(users(t), values("q=JO"))
	require.NoError(t, err)
	assert.Len(t, out.([]any), 2)
}

func TestApplyQuery_Sort(t *testing.T) {
	out, err := ApplyQuery(users(t), values("sort=id"))
	require.NoError(t, err)
	items := out.([]any)
	assert.Equal(t, json.Number("1"), items[0].(map[string]any)["id"])
	assert.Equal(t, json.Number("4"), items[3].(map[string]any)["id"])

	out, err = ApplyQuery(users(t), values("sort=-id"))
	require.NoError(t, err)
	items = out.([]any)
	assert.Equal(t, json.Number("4"), items[0].(map[string]any)["id"])

	out, err = ApplyQuery(users(t), values("sort=name,-id"))
	require.NoError(t, err)
	items = out.([]any)
	// Names: John(1), Jane(2), John(3), Zoe(4) -> sorted: Jane(2), John(3), John(1), Zoe(4).
	assert.Equal(t, json.Number("2"), items[0].(map[string]any)["id"])
	assert.Equal(t, json.Number("3"), items[1].(map[string]any)["id"])
	assert.Equal(t, json.Number("1"), items[2].(map[string]any)["id"])
}

func TestApplyQuery_Pagination(t *testing.T) {
	out, err := ApplyQuery(users(t), values("page=2&limit=2"))
	require.NoError(t, err)
	items := out.([]any)
	assert.Len(t, items, 2)
	assert.Equal(t, json.Number("3"), items[0].(map[string]any)["id"])

	out, err = ApplyQuery(users(t), values("page=10&limit=2"))
	require.NoError(t, err)
	assert.Len(t, out.([]any), 0)
}

func TestApplyQuery_Fields(t *testing.T) {
	out, err := ApplyQuery(users(t), values("fields=id,name"))
	require.NoError(t, err)
	items := out.([]any)
	obj := items[0].(map[string]any)
	assert.Equal(t, json.Number("1"), obj["id"])
	assert.Equal(t, "John", obj["name"])
	assert.NotContains(t, obj, "country")
}

func TestApplyQuery_InvalidPage(t *testing.T) {
	_, err := ApplyQuery(users(t), values("page=abc"))
	require.Error(t, err)

	_, err = ApplyQuery(users(t), values("page=0"))
	require.Error(t, err)

	_, err = ApplyQuery(users(t), values("limit=-1"))
	require.Error(t, err)
}

func TestApplyQuery_NonArrayReturnsUnchanged(t *testing.T) {
	obj := map[string]any{"version": "1.0"}
	out, err := ApplyQuery(obj, values("sort=version"))
	require.NoError(t, err)
	assert.Equal(t, obj, out)
}

func TestApplyQuery_PrimitiveRoot(t *testing.T) {
	out, err := ApplyQuery(json.Number("42"), values("page=1"))
	require.NoError(t, err)
	assert.Equal(t, json.Number("42"), out)
}
