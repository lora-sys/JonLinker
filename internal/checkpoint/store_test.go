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

func TestHasPrefixMatch(t *testing.T) {
	s := NewStore()
	ctx := context.Background()

	s.Set(ctx, "s123:profile", []byte(`{"name":"Alice"}`))
	s.Set(ctx, "s123:application_job123", []byte(`{}`))
	s.Set(ctx, "s456:profile", []byte(`{"name":"Bob"}`))

	if !s.HasPrefix(ctx, "s123:") {
		t.Fatal("HasPrefix(s123:) should be true")
	}
}

func TestHasPrefixNoMatch(t *testing.T) {
	s := NewStore()
	ctx := context.Background()

	s.Set(ctx, "s123:profile", []byte(`{}`))

	if s.HasPrefix(ctx, "s999:") {
		t.Fatal("HasPrefix(s999:) should be false")
	}
}

func TestHasPrefixEmpty(t *testing.T) {
	s := NewStore()
	ctx := context.Background()

	if s.HasPrefix(ctx, "") {
		t.Fatal("HasPrefix('') on empty store should be false")
	}
}

func TestHasPrefixEmptyPrefixOnData(t *testing.T) {
	s := NewStore()
	ctx := context.Background()

	s.Set(ctx, "anything", []byte(`{}`))

	if !s.HasPrefix(ctx, "") {
		t.Fatal("HasPrefix('') should return true when store has keys")
	}
}

func TestKeysMatch(t *testing.T) {
	s := NewStore()
	ctx := context.Background()

	s.Set(ctx, "s123:profile", []byte(`{}`))
	s.Set(ctx, "s123:application_job1", []byte(`{}`))
	s.Set(ctx, "s123:application_job2", []byte(`{}`))
	s.Set(ctx, "s456:profile", []byte(`{}`))

	keys := s.Keys(ctx, "s123:application_")
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d: %v", len(keys), keys)
	}
}

func TestKeysNoMatch(t *testing.T) {
	s := NewStore()
	ctx := context.Background()

	s.Set(ctx, "s123:profile", []byte(`{}`))

	keys := s.Keys(ctx, "s999:")
	if len(keys) != 0 {
		t.Fatalf("expected 0 keys, got %d", len(keys))
	}
}

func TestHasPrefixSessionApplication(t *testing.T) {
	s := NewStore()
	ctx := context.Background()

	s.Set(ctx, "s123:profile", []byte(`{"name":"Alice"}`))
	s.Set(ctx, "s456:application_", []byte(`{}`))

	if !s.HasPrefix(ctx, "s456:application_") {
		t.Fatal("HasPrefix(s456:application_) should be true")
	}
	if s.HasPrefix(ctx, "s789:") {
		t.Fatal("HasPrefix(s789:) should be false")
	}
}
