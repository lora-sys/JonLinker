package session

import (
	"context"
	"testing"
)

func TestStoreGetSet(t *testing.T) {
	s := NewStore()
	ctx := context.Background()

	type dummy struct {
		Name string
		Age  int
	}

	err := s.Set(ctx, "user1", dummy{Name: "Alice", Age: 30})
	if err != nil {
		t.Fatalf("Set: %v", err)
	}

	var got dummy
	err = s.Get(ctx, "user1", &got)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Name != "Alice" || got.Age != 30 {
		t.Fatalf("got %+v, want {Alice 30}", got)
	}
}

func TestStoreGetMissing(t *testing.T) {
	s := NewStore()
	ctx := context.Background()

	var v string
	err := s.Get(ctx, "nonexistent", &v)
	if err != nil {
		t.Fatalf("Get missing: %v", err)
	}
}

func TestStoreOverwrite(t *testing.T) {
	s := NewStore()
	ctx := context.Background()

	s.Set(ctx, "k", "v1")
	s.Set(ctx, "k", "v2")

	var got string
	s.Get(ctx, "k", &got)
	if got != "v2" {
		t.Fatalf("got %q, want v2", got)
	}
}
