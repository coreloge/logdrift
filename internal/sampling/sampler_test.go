package sampling_test

import (
	"testing"
	"time"

	"github.com/user/logdrift/internal/diff"
	"github.com/user/logdrift/internal/sampling"
)

func makeEntry(service string) diff.Entry {
	return diff.Entry{Service: service, Level: "info", Message: "test"}
}

func TestDefaultOptions_PassesAll(t *testing.T) {
	opts := sampling.DefaultOptions()
	if opts.Strategy != sampling.StrategyNone {
		t.Fatalf("expected StrategyNone, got %v", opts.Strategy)
	}
	if opts.Rate != 1.0 {
		t.Fatalf("expected rate 1.0, got %v", opts.Rate)
	}
}

func TestSampler_None_AcceptsAll(t *testing.T) {
	s := sampling.New(sampling.DefaultOptions())
	defer s.Stop()
	for i := 0; i < 100; i++ {
		if !s.Accept(makeEntry("svc")) {
			t.Fatal("StrategyNone should accept every entry")
		}
	}
}

func TestSampler_Random_ZeroRate_RejectsAll(t *testing.T) {
	s := sampling.New(sampling.Options{Strategy: sampling.StrategyRandom, Rate: 0.0})
	defer s.Stop()
	for i := 0; i < 200; i++ {
		if s.Accept(makeEntry("svc")) {
			t.Fatal("rate=0 should reject all entries")
		}
	}
}

func TestSampler_Random_FullRate_AcceptsAll(t *testing.T) {
	s := sampling.New(sampling.Options{Strategy: sampling.StrategyRandom, Rate: 1.0})
	defer s.Stop()
	for i := 0; i < 100; i++ {
		if !s.Accept(makeEntry("svc")) {
			t.Fatal("rate=1 should accept all entries")
		}
	}
}

func TestSampler_Random_PartialRate(t *testing.T) {
	s := sampling.New(sampling.Options{Strategy: sampling.StrategyRandom, Rate: 0.5})
	defer s.Stop()
	accepted := 0
	const n = 1000
	for i := 0; i < n; i++ {
		if s.Accept(makeEntry("svc")) {
			accepted++
		}
	}
	// Expect roughly 50 % ± 15 %
	if accepted < 350 || accepted > 650 {
		t.Fatalf("expected ~500 accepted out of %d, got %d", n, accepted)
	}
}

func TestSampler_RateLimit_EnforcesPerSecondLimit(t *testing.T) {
	const limit = 5
	s := sampling.New(sampling.Options{Strategy: sampling.StrategyRateLimit, Rate: limit})
	defer s.Stop()

	accepted := 0
	for i := 0; i < 20; i++ {
		if s.Accept(makeEntry("api")) {
			accepted++
		}
	}
	if accepted != limit {
		t.Fatalf("expected %d accepted, got %d", limit, accepted)
	}
}

func TestSampler_RateLimit_IndependentPerService(t *testing.T) {
	const limit = 3
	s := sampling.New(sampling.Options{Strategy: sampling.StrategyRateLimit, Rate: limit})
	defer s.Stop()

	for i := 0; i < 10; i++ {
		s.Accept(makeEntry("svcA"))
	}
	accepted := 0
	for i := 0; i < 10; i++ {
		if s.Accept(makeEntry("svcB")) {
			accepted++
		}
	}
	if accepted != limit {
		t.Fatalf("svcB counter should be independent; expected %d, got %d", limit, accepted)
	}
}

func TestSampler_Stop_DoesNotPanic(t *testing.T) {
	s := sampling.New(sampling.Options{Strategy: sampling.StrategyRateLimit, Rate: 10})
	time.Sleep(5 * time.Millisecond)
	s.Stop() // should not panic
}
