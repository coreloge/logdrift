package metrics

import (
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"
	"time"
)

// Snapshot holds a point-in-time copy of counter values.
type Snapshot struct {
	Timestamp  time.Time         `json:"timestamp"`
	Entries    map[string]int64  `json:"entries"`
	Drifts     map[string]int64  `json:"drifts"`
	TotalEntries int64           `json:"total_entries"`
	TotalDrifts  int64           `json:"total_drifts"`
}

// Exporter formats and writes Counter snapshots to an io.Writer.
type Exporter struct {
	counter *Counter
	format  string // "text" or "json"
}

// NewExporter creates an Exporter for the given Counter.
// format must be "text" or "json"; defaults to "text".
func NewExporter(c *Counter, format string) *Exporter {
	if format != "json" {
		format = "text"
	}
	return &Exporter{counter: c, format: format}
}

// Export writes the current counter state to w.
func (e *Exporter) Export(w io.Writer) error {
	snap := e.snapshot()
	if e.format == "json" {
		return e.writeJSON(w, snap)
	}
	return e.writeText(w, snap)
}

func (e *Exporter) snapshot() Snapshot {
	entries := e.counter.Entries()
	drifts := e.counter.Drifts()
	var te, td int64
	for _, v := range entries {
		te += v
	}
	for _, v := range drifts {
		td += v
	}
	return Snapshot{
		Timestamp:    time.Now().UTC(),
		Entries:      entries,
		Drifts:       drifts,
		TotalEntries: te,
		TotalDrifts:  td,
	}
}

func (e *Exporter) writeJSON(w io.Writer, s Snapshot) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(s)
}

func (e *Exporter) writeText(w io.Writer, s Snapshot) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintf(tw, "logdrift metrics @ %s\n", s.Timestamp.Format(time.RFC3339))
	fmt.Fprintf(tw, "SERVICE\tENTRIES\tDRIFTS\n")
	for svc, cnt := range s.Entries {
		fmt.Fprintf(tw, "%s\t%d\t%d\n", svc, cnt, s.Drifts[svc])
	}
	fmt.Fprintf(tw, "TOTAL\t%d\t%d\n", s.TotalEntries, s.TotalDrifts)
	return tw.Flush()
}
