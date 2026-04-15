// Package fingerprint computes a stable hash fingerprint for log entries,
// enabling deduplication, correlation, and caching by content identity.
package fingerprint

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"

	"github.com/user/logdrift/internal/diff"
)

// Options controls which fields contribute to the fingerprint.
type Options struct {
	// IncludeService includes the service label in the hash when true.
	IncludeService bool
	// IncludeLevel includes the log level in the hash when true.
	IncludeLevel bool
	// IncludeExtra includes all extra key/value fields in the hash when true.
	IncludeExtra bool
}

// DefaultOptions returns Options that hash message + level + service.
func DefaultOptions() Options {
	return Options{
		IncludeService: true,
		IncludeLevel:   true,
		IncludeExtra:   false,
	}
}

// Fingerprinter computes content fingerprints for log entries.
type Fingerprinter struct {
	opts Options
}

// New creates a Fingerprinter with the given Options.
func New(opts Options) *Fingerprinter {
	return &Fingerprinter{opts: opts}
}

// Compute returns a hex-encoded SHA-256 fingerprint for the entry.
func (f *Fingerprinter) Compute(e diff.Entry) string {
	h := sha256.New()

	fmt.Fprintf(h, "msg:%s", e.Message)

	if f.opts.IncludeLevel {
		fmt.Fprintf(h, "|lvl:%s", e.Level)
	}

	if f.opts.IncludeService {
		fmt.Fprintf(h, "|svc:%s", e.Service)
	}

	if f.opts.IncludeExtra && len(e.Extra) > 0 {
		keys := make([]string, 0, len(e.Extra))
		for k := range e.Extra {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Fprintf(h, "|%s:%s", k, e.Extra[k])
		}
	}

	return hex.EncodeToString(h.Sum(nil))
}
