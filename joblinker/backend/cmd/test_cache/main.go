package main

import (
	"context"
	"fmt"
	"time"

	"joblinker/internal/agent"
	"joblinker/internal/cache"
	"joblinker/internal/config"
)

func main() {
	toolCache := cache.NewToolCache(100, 5*time.Minute)
	exec := agent.NewToolExecutor(nil, nil, nil, nil, nil)
	exec.SetCache(toolCache)

	cfg, _ := config.ParseAgentToolConfig(`{"tools":{"query_jobs":true,"create_offer":false,"get_candidate":true}}`)
	exec.SetAllowedTools(cfg)

	result, err := exec.ExecuteTool(context.Background(), [16]byte{}, "query_jobs", map[string]interface{}{
		"location": "SF",
		"skills":   []interface{}{"golang", "python"},
	})
	if err != nil {
		fmt.Printf("query_jobs error: %v\n", err)
	} else {
		fmt.Printf("query_jobs success: %+v\n", result)
	}

	result2, err2 := exec.ExecuteTool(context.Background(), [16]byte{}, "create_offer", map[string]interface{}{
		"match_id": "test-match",
		"salary":   150000,
	})
	if err2 != nil {
		fmt.Printf("create_offer correctly denied: %v\n", err2)
	} else {
		fmt.Printf("create_offer incorrectly allowed: %+v\n", result2)
	}

	stats := toolCache.Stats()
	fmt.Printf("\nCache stats: %+v\n", stats)

	fmt.Printf("\nCanUseTool tests:\n")
	fmt.Printf("  query_jobs: %v (expected: true)\n", cfg.CanUseTool("query_jobs"))
	fmt.Printf("  create_offer: %v (expected: false)\n", cfg.CanUseTool("create_offer"))
}