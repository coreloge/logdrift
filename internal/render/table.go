package render

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/user/logdrift/internal/diff"
	"github.com/user/logdrift/internal/snapshot"
)

// TableRenderer writes a side-by-side diff summary table for a snapshot
// comparison result to the supplied writer.
type TableRenderer struct {
	w io.Writer
	tw *tabwriter.Writer
}

// NewTableRenderer creates a TableRenderer that writes to w.
func NewTableRenderer(w io.Writer) *TableRenderer {
	return &TableRenderer{
		w:  w,
		tw: tabwriter.NewWriter(w, 0, 0, 2, ' ', 0),
	}
}

// RenderSnapshot prints a formatted comparison table derived from a snapshot
// diff result. Each row represents a service pair that was compared.
func (r *TableRenderer) RenderSnapshot(result snapshot.CompareResult) error {
	fmt.Fprintln(r.tw, "SERVICE A\tSERVICE B\tSTATUS\tDELTAS")
	fmt.Fprintln(r.tw, strings.Repeat("-", 10)+"\t"+strings.Repeat("-", 10)+"\t"+strings.Repeat("-", 8)+"\t"+strings.Repeat("-", 30))

	for _, row := range result.Rows {
		status := "MATCH"
		deltas := "-"
		if len(row.Deltas) > 0 {
			status = "DIFF"
			parts := make([]string, 0, len(row.Deltas))
			for _, d := range row.Deltas {
				parts = append(parts, diff.FormatDelta(d))
			}
			deltas = strings.Join(parts, "; ")
		}
		fmt.Fprintf(r.tw, "%s\t%s\t%s\t%s\n", row.ServiceA, row.ServiceB, status, deltas)
	}

	return r.tw.Flush()
}
