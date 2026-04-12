package checkpoint_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yourorg/logdrift/internal/checkpoint"
)

func tempPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "checkpoints.json")
}

func TestNew_CreatesEmptyStore(t *testing.T) {
	s, err := checkpoint.New(tempPath(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := s.Services(); len(got) != 0 {
		t.Fatalf("expected empty store, got %v", got)
	}
}

func TestSet_And_Get(t *testing.T) {
	s, _ := checkpoint.New(tempPath(t))
	if err := s.Set("svc-a", 1024); err != nil {
		t.Fatalf("Set: %v", err)
	}
	got, err := s.Get("svc-a")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != 1024 {
		t.Fatalf("expected 1024, got %d", got)
	}
}

func TestGet_MissingKey_ReturnsErrNotFound(t *testing.T) {
	s, _ := checkpoint.New(tempPath(t))
	_, err := s.Get("missing")
	if err != checkpoint.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestDelete_RemovesEntry(t *testing.T) {
	s, _ := checkpoint.New(tempPath(t))
	_ = s.Set("svc-b", 512)
	_ = s.Delete("svc-b")
	_, err := s.Get("svc-b")
	if err != checkpoint.ErrNotFound {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestPersistence_ReloadsFromDisk(t *testing.T) {
	p := tempPath(t)
	s1, _ := checkpoint.New(p)
	_ = s1.Set("svc-c", 9999)

	s2, err := checkpoint.New(p)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	got, err := s2.Get("svc-c")
	if err != nil {
		t.Fatalf("Get after reload: %v", err)
	}
	if got != 9999 {
		t.Fatalf("expected 9999, got %d", got)
	}
}

func TestNew_InvalidJSON_ReturnsError(t *testing.T) {
	p := tempPath(t)
	_ = os.WriteFile(p, []byte("not-json"), 0o644)
	_, err := checkpoint.New(p)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
