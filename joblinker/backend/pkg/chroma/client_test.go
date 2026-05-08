//go:build integration
// +build integration

package chroma

import (
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestChromaClient_NewClient(t *testing.T) {
	c := NewClient("localhost", 8000)
	if c == nil {
		t.Fatal("NewClient returned nil")
	}
	if c.host != "localhost" {
		t.Errorf("expected host localhost, got %s", c.host)
	}
	if c.port != 8000 {
		t.Errorf("expected port 8000, got %d", c.port)
	}
}

func TestChromaClient_NewClient_Defaults(t *testing.T) {
	c := NewClient("", 0)
	if c.host != "localhost" {
		t.Errorf("expected default host localhost, got %s", c.host)
	}
	if c.port != 8000 {
		t.Errorf("expected default port 8000, got %d", c.port)
	}
}

func TestChromaClient_Heartbeat(t *testing.T) {
	chromaHost := getEnv("CHROMA_HOST", "localhost")
	chromaPort := getEnvInt("CHROMA_PORT", 8000)

	c := NewClient(chromaHost, chromaPort)
	if err := c.Heartbeat(); err != nil {
		t.Fatalf("Chroma heartbeat failed at %s:%d: %v", chromaHost, chromaPort, err)
	}
}

func TestChromaClient_CollectionCRUD(t *testing.T) {
	chromaHost := getEnv("CHROMA_HOST", "localhost")
	chromaPort := getEnvInt("CHROMA_PORT", 8000)
	c := NewClient(chromaHost, chromaPort)

	// Create/Get collection
	colName := "test_crud_" + randomID()
	col, err := c.GetOrCreateCollection(colName)
	if err != nil {
		t.Fatalf("GetOrCreateCollection failed: %v", err)
	}
	if col.Name != colName {
		t.Errorf("expected collection name %s, got %s", colName, col.Name)
	}

	// Add documents
	testID := "doc1_" + randomID()
	err = c.Add(col.Name, []string{testID}, []string{"Go developer with Kubernetes experience"}, []map[string]interface{}{{"skill": "go", "domain": "backend"}})
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Query — should find the Go developer
	result, err := c.Query(col.Name, []string{"Golang programmer"}, 3, nil)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if len(result.IDs) == 0 || len(result.IDs[0]) == 0 {
		t.Error("Query returned no results")
	}
	found := false
	for _, ids := range result.IDs {
		for _, id := range ids {
			if id == testID {
				found = true
				break
			}
		}
	}
	if !found {
		t.Logf("Query result IDs: %v", result.IDs)
		t.Error("semantic search did not return expected document")
	}

	// Delete
	err = c.Delete(col.Name, []string{testID})
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deleted
	docs, err := c.Get(col.Name, []string{testID}, []string{"documents"})
	if err != nil {
		t.Fatalf("Get after delete failed: %v", err)
	}
	if len(docs) != 0 {
		t.Error("document still exists after delete")
	}
}

func TestChromaClient_AddAndQuery(t *testing.T) {
	chromaHost := getEnv("CHROMA_HOST", "localhost")
	chromaPort := getEnvInt("CHROMA_PORT", 8000)
	c := NewClient(chromaHost, chromaPort)

	col, _ := c.GetOrCreateCollection("test_search_" + randomID())
	colName := col.Name

	// Add multiple docs
	ids := []string{"doc1", "doc2", "doc3"}
	docs := []string{
		"Python developer with 5 years of machine learning experience",
		"Java backend engineer specializing in Spring Boot microservices",
		"Frontend React developer with TypeScript and CSS expertise",
	}
	metas := []map[string]interface{}{
		{"language": "python", "domain": "ml"},
		{"language": "java", "domain": "backend"},
		{"language": "typescript", "domain": "frontend"},
	}
	if err := c.Add(colName, ids, docs, metas); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Query for ML engineer — should match doc1
	result, err := c.Query(colName, []string{"machine learning engineer python"}, 2, nil)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if len(result.IDs) == 0 || len(result.IDs[0]) == 0 {
		t.Fatal("Query returned no results")
	}

	// doc1 should be top result (highest similarity)
	if len(result.IDs[0]) == 0 || result.IDs[0][0] != "doc1" {
		t.Errorf("expected doc1 as top result for ML query, got %v", result.IDs[0])
	}

	// Clean up
	c.Delete(colName, ids)
}

func TestChromaClient_EnsureCollection(t *testing.T) {
	chromaHost := getEnv("CHROMA_HOST", "localhost")
	chromaPort := getEnvInt("CHROMA_PORT", 8000)
	c := NewClient(chromaHost, chromaPort)

	collName := "test_ensure_" + randomID()

	// First call should create and return a UUID
	uuid1, err := c.EnsureCollection(collName)
	if err != nil {
		t.Fatalf("EnsureCollection first call failed: %v", err)
	}
	if uuid1 == "" {
		t.Fatal("EnsureCollection returned empty UUID")
	}

	// Second call should return the same UUID (cached)
	uuid2, err := c.EnsureCollection(collName)
	if err != nil {
		t.Fatalf("EnsureCollection second call failed: %v", err)
	}
	if uuid1 != uuid2 {
		t.Errorf("EnsureCollection returned different UUIDs: %s vs %s", uuid1, uuid2)
	}
}

// Helper functions
func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultVal
}

func randomID() string {
	return fmt.Sprintf("test-%d", time.Now().UnixNano())
}
