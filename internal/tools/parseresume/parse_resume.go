package parseresume

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type Tool struct{}

func NewTool(_ string) *Tool {
	return &Tool{}
}

func (t *Tool) ParseFile(_ context.Context, filename string, data []byte) (string, error) {
	dir, err := os.MkdirTemp("", "resume-*")
	if err != nil {
		return "", fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(dir)

	tmpPath := filepath.Join(dir, filepath.Base(filename))
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return "", fmt.Errorf("write temp file: %w", err)
	}

	outPath := tmpPath + ".md"
	cmd := exec.Command("firecrawl", "parse", tmpPath, "-o", outPath, "--timeout", "30000")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("firecrawl parse failed: %w\n%s", err, string(output))
	}

	parsed, err := os.ReadFile(outPath)
	if err != nil {
		return "", fmt.Errorf("read output: %w", err)
	}

	return string(parsed), nil
}
