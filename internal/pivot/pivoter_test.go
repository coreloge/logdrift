package pivot_test

import (
	"context"
	"testing"

	"github.com/yourorg/logdrift/internal/diff"
	"github.com/yourorg/logdrift/internal/pivot"
)

func makeEntry(svc, level, msg string, extra map[string]string) diff.Entry {
	return diff.Entry{Service: svc, Level: level, Message: msg, Extra: extra}
}

func TestNew_EmptyField_ReturnsError(t *testing.T) {
	_, err := pivot.New(pivot.Options{Field: ""})
	if err == nil {
		t.Fatal("expected error for empty field")
	}
}

func TestNew_ValidOptions_NoError(t *testing.T) {
	_, err := pivot.New(pivot.DefaultOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRecord_ByLevel_CountsCorrectly(t *testing.T) {
	p, _ := pivot.New(pivot.DefaultOptions())
	p.Record(makeEntry("svcA", "error", "boom", nil))
	p.Record(makeEntry("svcA", "error", "boom2", nil))
	p.Record(makeEntry("svcB", "error", "oops", nil))
	p.Record(makeEntry("svcA", "info", "ok", nil))

	tbl := p.Table()
	if tbl.Rows["error"]["svcA"] != 2 {
		t.Errorf("expected 2, got %d", tbl.Rows["error"]["svcA"])
	}
	if tbl.Rows["error"]["svcB"] != 1 {
		t.Errorf("expected 1, got %d", tbl.Rows["error"]["svcB"])
	}
	if tbl.Rows["info"]["svcA"] != 1 {
		t.Errorf("expected 1, got %d", tbl.Rows["info"]["svcA"])
	}
}

func TestRecord_CustomField_SkipsMissingField(t *testing.T) {
	p, _ := pivot.New(pivot.Options{Field: "status"})
	p.Record(makeEntry("svcA", "info", "ok", map[string]string{"status": "200"}))
	p.Record(makeEntry("svcB", "info", "ok", nil)) // no status field

	tbl := p.Table()
	if _, ok := tbl.Rows["200"]; !ok {
		t.Fatal("expected row for status=200")
	}
	if tbl.Rows["200"]["svcB"] != 0 {
		t.Errorf("svcB should not appear under status=200")
	}
}

func TestReset_ClearsData(t *testing.T) {
	p, _ := pivot.New(pivot.DefaultOptions())
	p.Record(makeEntry("svcA", "error", "boom", nil))
	p.Reset()
	tbl := p.Table()
	if len(tbl.Rows) != 0 {
		t.Errorf("expected empty table after reset, got %d rows", len(tbl.Rows))
	}
}

func TestTable_ReturnsCopy(t *testing.T) {
	p, _ := pivot.New(pivot.DefaultOptions())
	p.Record(makeEntry("svcA", "info", "hi", nil))
	t1 := p.Table()
	t1.Rows["info"]["svcA"] = 999
	t2 := p.Table()
	if t2.Rows["info"]["svcA"] == 999 {
		t.Error("Table should return an independent copy")
	}
}

func TestStream_ForwardsAllEntries(t *testing.T) {
	p, _ := pivot.New(pivot.DefaultOptions())
	in := make(chan diff.Entry, 3)
	in <- makeEntry("svcA", "info", "a", nil)
	in <- makeEntry("svcB", "warn", "b", nil)
	in <- makeEntry("svcA", "error", "c", nil)
	close(in)

	ctx := context.Background()
	out := pivot.Stream(ctx, p, in)

	var got []diff.Entry
	for e := range out {
		got = append(got, e)
	}
	if len(got) != 3 {
		t.Errorf("expected 3 forwarded entries, got %d", len(got))
	}
	tbl := p.Table()
	if tbl.Rows["error"]["svcA"] != 1 {
		t.Errorf("expected pivot to record error for svcA")
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	p, _ := pivot.New(pivot.DefaultOptions())
	in := make(chan diff.Entry)
	ctx, cancel := context.WithCancel(context.Background())
	out := pivot.Stream(ctx, p, in)
	cancel()
	for range out {
	}
}
