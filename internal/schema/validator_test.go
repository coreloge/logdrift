package schema_test

import (
	"context"
	"testing"
	"time"

	"github.com/user/logdrift/internal/diff"
	"github.com/user/logdrift/internal/schema"
)

func makeEntry(level, message string, fields map[string]string) diff.Entry {
	if fields == nil {
		fields = map[string]string{}
	}
	return diff.Entry{Service: "svc", Level: level, Message: message, Fields: fields}
}

func TestNew_InvalidPattern_ReturnsError(t *testing.T) {
	opts := schema.DefaultOptions()
	opts.Rules["level"] = schema.FieldRule{Pattern: "[invalid"}
	_, err := schema.New(opts)
	if err == nil {
		t.Fatal("expected error for invalid regex pattern")
	}
}

func TestValidate_RequiredField_Present(t *testing.T) {
	opts := schema.DefaultOptions()
	opts.Rules["level"] = schema.FieldRule{Required: true}
	v, _ := schema.New(opts)
	entry := makeEntry("info", "hello", nil)
	if err := v.Validate(entry); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidate_RequiredField_Missing(t *testing.T) {
	opts := schema.DefaultOptions()
	opts.Rules["level"] = schema.FieldRule{Required: true}
	v, _ := schema.New(opts)
	entry := makeEntry("", "hello", nil)
	if err := v.Validate(entry); err == nil {
		t.Fatal("expected error for missing required field")
	}
}

func TestValidate_PatternMatch_Passes(t *testing.T) {
	opts := schema.DefaultOptions()
	opts.Rules["level"] = schema.FieldRule{Pattern: "^(info|warn|error)$"}
	v, _ := schema.New(opts)
	entry := makeEntry("warn", "something", nil)
	if err := v.Validate(entry); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidate_PatternMatch_Fails(t *testing.T) {
	opts := schema.DefaultOptions()
	opts.Rules["level"] = schema.FieldRule{Pattern: "^(info|warn|error)$"}
	v, _ := schema.New(opts)
	entry := makeEntry("DEBUG", "something", nil)
	if err := v.Validate(entry); err == nil {
		t.Fatal("expected error for pattern mismatch")
	}
}

func TestStream_DropInvalid_FiltersEntries(t *testing.T) {
	opts := schema.DefaultOptions()
	opts.Rules["level"] = schema.FieldRule{Required: true}
	opts.DropInvalid = true
	v, _ := schema.New(opts)

	in := make(chan diff.Entry, 3)
	in <- makeEntry("", "no level", nil)
	in <- makeEntry("info", "valid", nil)
	in <- makeEntry("", "also bad", nil)
	close(in)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	out := schema.Stream(ctx, v, in)
	var results []diff.Entry
	for e := range out {
		results = append(results, e)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 valid entry, got %d", len(results))
	}
	if results[0].Level != "info" {
		t.Errorf("unexpected entry level: %s", results[0].Level)
	}
}

func TestStream_StopsOnContextCancel(t *testing.T) {
	opts := schema.DefaultOptions()
	v, _ := schema.New(opts)

	in := make(chan diff.Entry)
	ctx, cancel := context.WithCancel(context.Background())
	out := schema.Stream(ctx, v, in)
	cancel()

	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("expected channel to be closed after cancel")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for stream to stop")
	}
}
