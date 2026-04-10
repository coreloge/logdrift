package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yourorg/logdrift/internal/config"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "logdrift-*.yaml")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_Valid(t *testing.T) {
	path := writeTemp(t, `
services:
  - name: api
    command: "docker logs -f api"
    color: cyan
  - name: worker
    file: "/var/log/worker.log"
diff_mode: true
max_lines: 200
`)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Services) != 2 {
		t.Errorf("expected 2 services, got %d", len(cfg.Services))
	}
	if cfg.MaxLines != 200 {
		t.Errorf("expected max_lines=200, got %d", cfg.MaxLines)
	}
	if !cfg.DiffMode {
		t.Error("expected diff_mode=true")
	}
}

func TestLoad_DefaultMaxLines(t *testing.T) {
	path := writeTemp(t, `
services:
  - name: svc
    command: "tail -f /tmp/test.log"
`)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.MaxLines != config.DefaultMaxLines {
		t.Errorf("expected default max_lines=%d, got %d", config.DefaultMaxLines, cfg.MaxLines)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := config.Load(filepath.Join(t.TempDir(), "nonexistent.yaml"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoad_NoServices(t *testing.T) {
	path := writeTemp(t, `services: []\n`)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected validation error for empty services")
	}
}

func TestLoad_DuplicateServiceName(t *testing.T) {
	path := writeTemp(t, `
services:
  - name: api
    command: "echo hello"
  - name: api
    command: "echo world"
`)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error for duplicate service name")
	}
}

func TestLoad_MissingCommandAndFile(t *testing.T) {
	path := writeTemp(t, `
services:
  - name: broken
`)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error when neither command nor file is set")
	}
}
