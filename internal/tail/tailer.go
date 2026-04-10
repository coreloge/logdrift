package tail

import (
	"bufio"
	"context"
	"io"
	"os"
	"time"
)

// Line represents a single log line with metadata.
type Line struct {
	Service string
	Text    string
	Time    time.Time
}

// Tailer tails a log file and emits lines to a channel.
type Tailer struct {
	Service  string
	FilePath string
	Lines    chan Line
}

// New creates a new Tailer for the given service and file path.
func New(service, filePath string) *Tailer {
	return &Tailer{
		Service:  service,
		FilePath: filePath,
		Lines:    make(chan Line, 64),
	}
}

// Tail opens the file and streams new lines into t.Lines until ctx is cancelled.
// It seeks to the end of the file before starting to avoid replaying history.
func (t *Tailer) Tail(ctx context.Context) error {
	f, err := os.Open(t.FilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Seek to end so we only tail new content.
	if _, err := f.Seek(0, io.SeekEnd); err != nil {
		return err
	}

	reader := bufio.NewReader(f)

	for {
		select {
		case <-ctx.Done():
			close(t.Lines)
			return ctx.Err()
		default:
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// No new data yet; wait briefly before retrying.
				time.Sleep(100 * time.Millisecond)
				continue
			}
			return err
		}

		if len(line) > 0 {
			t.Lines <- Line{
				Service: t.Service,
				Text:    line,
				Time:    time.Now(),
			}
		}
	}
}
