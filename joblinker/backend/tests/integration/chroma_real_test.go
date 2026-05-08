//go:build integration
// +build integration

// Real integration tests for Chroma collection resolution and persistence.
package integration

import (
	"fmt"
	"testing"
	"time"

	"joblinker/pkg/chroma"
	"joblinker/tests/testutil"
)

// TestChromaEnsureCollection verifies that EnsureCollection returns a non-empty UUID
// and subsequent calls return the same UUID (caching works).
func TestChromaEnsureCollection(t *testing.T) {
	chromaHost := testutil.RequireEnv(t, "CHROMA_HOST")
	chromaPort := 8000
	c := chroma.NewClient(chromaHost, chromaPort)

	if err := c.Heartbeat(); err != nil {
		t.Fatalf("Chroma not reachable: %v — %s", err, testutil.ENV_MISSING_MSG)
	}

	collName := fmt.Sprintf("test_ensure_%d", time.Now().UnixNano())

	uuid1, err := c.EnsureCollection(collName)
	if err != nil {
		t.Fatalf("EnsureCollection first call failed: %v", err)
	}
	if uuid1 == "" {
		t.Fatal("EnsureCollection returned empty UUID")
	}

	uuid2, err := c.EnsureCollection(collName)
	if err != nil {
		t.Fatalf("EnsureCollection second call failed: %v", err)
	}
	if uuid1 != uuid2 {
		t.Errorf("EnsureCollection returned different UUIDs: %s vs %s", uuid1, uuid2)
	}
}

// TestChromaCollectionPersistence verifies that documents stored via Add
// can be retrieved via Query using collection names (not UUIDs).
func TestChromaCollectionPersistence(t *testing.T) {
	chromaHost := testutil.RequireEnv(t, "CHROMA_HOST")
	chromaPort := 8000
	c := chroma.NewClient(chromaHost, chromaPort)

	if err := c.Heartbeat(); err != nil {
		t.Fatalf("Chroma not reachable: %v — %s", err, testutil.ENV_MISSING_MSG)
	}

	collName := fmt.Sprintf("test_persist_%d", time.Now().UnixNano())

	// Store a document using collection name
	docID := fmt.Sprintf("doc_%d", time.Now().UnixNano())
	err := c.Add(collName, []string{docID}, []string{"Go developer with Kubernetes experience"}, []map[string]interface{}{{"skill": "go"}})
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Query using collection name — should find the document
	result, err := c.Query(collName, []string{"Golang programmer"}, 3, nil)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if len(result.IDs) == 0 || len(result.IDs[0]) == 0 {
		t.Fatal("Query returned no results after Add")
	}

	found := false
	for _, ids := range result.IDs {
		for _, id := range ids {
			if id == docID {
				found = true
				break
			}
		}
	}
	if !found {
		t.Errorf("Query did not return stored document %s, got IDs: %v", docID, result.IDs)
	}

	// Clean up
	if err := c.Delete(collName, []string{docID}); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}
