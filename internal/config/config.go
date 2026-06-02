package config

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	OpenAIBaseURL string
	OpenAIAPIKey  string
	OpenAIModel   string
	FirecrawlKey  string
	ServerPort    string
	FrontendURL   string
	SessionFile   string
	RedisURL      string
	MemoryBackend string
}

func Load() *Config {
	loadEnvFile()

	apiKey := os.Getenv("OPENAI_API_KEY")
	firecrawlKey := os.Getenv("FIRECRAWL_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY is required")
	}
	if firecrawlKey == "" {
		log.Fatal("FIRECRAWL_API_KEY is required")
	}

	return &Config{
		OpenAIBaseURL: normalizeBaseURL(os.Getenv("OPENAI_BASE_URL")),
		OpenAIAPIKey:  apiKey,
		OpenAIModel:   os.Getenv("OPENAI_MODEL"),
		FirecrawlKey:  firecrawlKey,
		ServerPort:    envDefault("SERVER_PORT", "8080"),
		FrontendURL:   envDefault("FRONTEND_URL", "http://localhost:3000"),
		SessionFile:   os.Getenv("SESSION_FILE"),
		RedisURL:      os.Getenv("REDIS_URL"),
		MemoryBackend: envDefault("MEMORY_BACKEND", "inmem"),
	}
}

func envDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func normalizeBaseURL(url string) string {
	if url == "" {
		return ""
	}
	url = strings.TrimRight(url, "/")
	if !strings.HasSuffix(url, "/v1") {
		url += "/v1"
	}
	return url
}

func loadEnvFile() {
	dir := findRootDir()
	for _, name := range []string{".env", ".env.local"} {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); err == nil {
			_ = godotenv.Load(p)
			return
		}
	}
}

func findRootDir() string {
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return dir
		}
		dir = parent
	}
}
