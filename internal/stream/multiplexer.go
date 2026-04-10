package stream

import (
	"sync"

	"github.com/user/logdrift/internal/tail"
)

// Entry represents a log line from a named service.
type Entry struct {
	Service string
	Line    string
}

// Multiplexer fans in log lines from multiple tailers into a single channel.
type Multiplexer struct {
	services map[string]*tail.Tailer
	out      chan Entry
	stop     chan struct{}
	wg       sync.WaitGroup
}

// New creates a Multiplexer from a map of service name -> file path.
func New(sources map[string]string) (*Multiplexer, error) {
	services := make(map[string]*tail.Tailer, len(sources))
	for name, path := range sources {
		t, err := tail.New(path)
		if err != nil {
			return nil, err
		}
		services[name] = t
	}
	return &Multiplexer{
		services: services,
		out:      make(chan Entry, 64),
		stop:     make(chan struct{}),
	}, nil
}

// Start begins tailing all sources and multiplexing lines into Out().
func (m *Multiplexer) Start() {
	for name, t := range m.services {
		m.wg.Add(1)
		go func(svc string, tailer *tail.Tailer) {
			defer m.wg.Done()
			lines := tailer.Tail(m.stop)
			for line := range lines {
				select {
				case m.out <- Entry{Service: svc, Line: line}:
				case <-m.stop:
					return
				}
			}
		}(name, t)
	}
}

// Out returns the channel of multiplexed log entries.
func (m *Multiplexer) Out() <-chan Entry {
	return m.out
}

// Stop signals all tailers to stop and waits for goroutines to finish.
func (m *Multiplexer) Stop() {
	close(m.stop)
	m.wg.Wait()
	close(m.out)
}
