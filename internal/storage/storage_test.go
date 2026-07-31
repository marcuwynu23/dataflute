package storage

import (
	"testing"

	"github.com/marcuwynu23/dataflute/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorage_UpsertGetDelete(t *testing.T) {
	s := New()
	d := &models.Dataset{Name: "users", Data: []any{"a"}}

	s.Upsert("users", d)
	got, ok := s.Get("users")
	require.True(t, ok)
	assert.Equal(t, "users", got.Name)

	s.Delete("users")
	_, ok = s.Get("users")
	assert.False(t, ok)
}

func TestStorage_NormalizesNames(t *testing.T) {
	s := New()
	d := &models.Dataset{Name: "users", Data: []any{}}

	s.Upsert("/users", d)
	_, ok := s.Get("users")
	assert.True(t, ok)
	_, ok = s.Get("/users")
	assert.True(t, ok)

	got, _ := s.Get("users")
	assert.Equal(t, "users", got.Name)
}

func TestStorage_NamesSorted(t *testing.T) {
	s := New()
	for _, name := range []string{"zebra", "apple", "mango"} {
		s.Upsert(name, &models.Dataset{Name: name, Data: nil})
	}
	assert.Equal(t, []string{"apple", "mango", "zebra"}, s.Names())
	assert.Equal(t, 3, s.Count())
}

func TestStorage_ConcurrentAccess(t *testing.T) {
	s := New()
	done := make(chan struct{})
	for i := 0; i < 8; i++ {
		go func(n int) {
			defer func() { done <- struct{}{} }()
			for j := 0; j < 100; j++ {
				name := "ds" + string(rune(n%3))
				s.Upsert(name, &models.Dataset{Name: name, Data: nil})
				_, _ = s.Get(name)
				_ = s.Names()
				_ = s.Count()
			}
		}(i)
	}
	for i := 0; i < 8; i++ {
		<-done
	}
	assert.Equal(t, 3, s.Count())
}
