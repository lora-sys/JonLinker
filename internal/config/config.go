package config

import (
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
}

func Load() *Config {
	loadEnvFile()
	return &Config{
		OpenAIBaseURL: normalizeBaseURL(os.Getenv("OPENAI_BASE_URL")),
		OpenAIAPIKey:  os.Getenv("OPENAI_API_KEY"),
		OpenAIModel:   os.Getenv("OPENAI_MODEL"),
		FirecrawlKey:  os.Getenv("FIRECRAWL_API_KEY"),
		ServerPort:    envDefault("SERVER_PORT", "8080"),
		FrontendURL:   envDefault("FRONTEND_URL", "http://localhost:3000"),
		SessionFile:   os.Getenv("SESSION_FILE"),
	}
}

func envDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func normalizeBaseURL(url string) string {
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
