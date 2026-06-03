package memory

import (
	"log"

	"github.com/lora-sys/JonLinker/internal/config"
	"github.com/redis/go-redis/v9"
)

func NewStore(cfg *config.Config) MemoryStore {
	switch cfg.MemoryBackend {
	case "redis":
		if cfg.RedisURL == "" {
			log.Fatal("MEMORY_BACKEND=redis requires REDIS_URL")
		}
		opts, err := redis.ParseURL(cfg.RedisURL)
		if err != nil {
			log.Fatalf("parse REDIS_URL: %v", err)
		}
		cli := redis.NewClient(opts)
		log.Printf("memory: using Redis (%s)", opts.Addr)
		return NewRedisStore(cli)

	case "file":
		if cfg.SessionFile == "" {
			log.Fatal("MEMORY_BACKEND=file requires SESSION_FILE")
		}
		log.Printf("memory: using file (%s)", cfg.SessionFile)
		return NewFileStore(cfg.SessionFile)

	default:
		log.Println("memory: using in-memory (no persistence)")
		return NewInMemoryStore()
	}
}
