package checkpoint

import (
	"context"
	"encoding/json"
	"testing"
)

func TestStoreGetSet(t *testing.T) {
	s := NewStore()
	ctx := context.Background()

	type dummy struct {
		Name string
		Age  int
	}

	want := dummy{Name: "Alice", Age: 30}
	b, _ := json.Marshal(want)
	err := s.Set(ctx, "user1", b)
	if err != nil {
		t.Fatalf("Set: %v", err)
	}

	got, ok, err := s.Get(ctx, "user1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !ok {
		t.Fatal("Get: not found")
	}
	var gotV dummy
	json.Unmarshal(got, &gotV)
	if gotV.Name != "Alice" || gotV.Age != 30 {
		t.Fatalf("got %+v, want {Alice 30}", gotV)
	}
}

func TestStoreGetMissing(t *testing.T) {
	s := NewStore()
	ctx := context.Background()

	_, ok, err := s.Get(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("Get missing: %v", err)
	}
	if ok {
		t.Fatal("Get missing: expected not found")
	}
}

func TestStoreOverwrite(t *testing.T) {
	s := NewStore()
	ctx := context.Background()

	s.Set(ctx, "k", []byte("v1"))
	s.Set(ctx, "k", []byte("v2"))

	got, _, _ := s.Get(ctx, "k")
	if string(got) != "v2" {
		t.Fatalf("got %q, want v2", string(got))
	}
}
