package parser

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeParser is a minimal parser used to exercise the registry.
type fakeParser struct {
	exts []string
}

func (f *fakeParser) Extensions() []string      { return f.exts }
func (f *fakeParser) Parse(string) (any, error) { return "fake", nil }

func TestRegistry_RegisterAndLookup(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&fakeParser{exts: []string{".json", "YAML"}})

	p, ok := reg.ParserFor("data.json")
	require.True(t, ok)
	assert.NotNil(t, p)

	p, ok = reg.ParserFor("data.yaml")
	require.True(t, ok)
	assert.NotNil(t, p)

	p, ok = reg.ParserFor("data.yml")
	assert.False(t, ok)
	assert.Nil(t, p)
}

func TestRegistry_Extensions(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&fakeParser{exts: []string{".json"}})
	reg.Register(&fakeParser{exts: []string{".csv", ".tsv"}})

	assert.Equal(t, []string{".csv", ".json", ".tsv"}, reg.Extensions())
}

func TestRegistry_NoParser(t *testing.T) {
	reg := NewRegistry()
	p, ok := reg.ParserFor(filepath.Join(t.TempDir(), "file.xyz"))
	assert.False(t, ok)
	assert.Nil(t, p)
}

func TestRegistry_ConcurrentUse(t *testing.T) {
	reg := NewRegistry()
	done := make(chan struct{})
	for i := 0; i < 8; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			for j := 0; j < 100; j++ {
				reg.Register(&fakeParser{exts: []string{".json"}})
				_, _ = reg.ParserFor("x.json")
				_ = reg.Extensions()
			}
		}()
	}
	for i := 0; i < 8; i++ {
		<-done
	}
}
