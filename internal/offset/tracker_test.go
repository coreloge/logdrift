package offset_test

import (
	"sort"
	"testing"

	"github.com/logdrift/logdrift/internal/offset"
)

func TestNew_StartsEmpty(t *testing.T) {
	tr := offset.New()
	if got := tr.Services(); len(got) != 0 {
		t.Fatalf("expected empty tracker, got %v", got)
	}
}

func TestSet_And_Get(t *testing.T) {
	tr := offset.New()
	tr.Set("api", 1024)

	v, err := tr.Get("api")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != 1024 {
		t.Fatalf("expected 1024, got %d", v)
	}
}

func TestGet_MissingService_ReturnsErrNotFound(t *testing.T) {
	tr := offset.New()
	_, err := tr.Get("missing")
	if err != offset.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestSet_Overwrites(t *testing.T) {
	tr := offset.New()
	tr.Set("svc", 100)
	tr.Set("svc", 999)

	v, err := tr.Get("svc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != 999 {
		t.Fatalf("expected 999, got %d", v)
	}
}

func TestDelete_RemovesEntry(t *testing.T) {
	tr := offset.New()
	tr.Set("svc", 42)
	tr.Delete("svc")

	_, err := tr.Get("svc")
	if err != offset.ErrNotFound {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestServices_ReturnsAllKeys(t *testing.T) {
	tr := offset.New()
	tr.Set("alpha", 1)
	tr.Set("beta", 2)
	tr.Set("gamma", 3)

	svcs := tr.Services()
	sort.Strings(svcs)

	expected := []string{"alpha", "beta", "gamma"}
	for i, e := range expected {
		if svcs[i] != e {
			t.Fatalf("expected %v, got %v", expected, svcs)
		}
	}
}

func TestReset_ClearsAll(t *testing.T) {
	tr := offset.New()
	tr.Set("a", 1)
	tr.Set("b", 2)
	tr.Reset()

	if got := tr.Services(); len(got) != 0 {
		t.Fatalf("expected empty after reset, got %v", got)
	}
}
