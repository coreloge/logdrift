package stream

import "strings"

// FilterOptions controls which entries are forwarded.
type FilterOptions struct {
	// Levels is an allowlist of log levels (e.g. ["error","warn"]).
	// Empty means all levels are accepted.
	Levels []string
	// Services is an allowlist of service names.
	// Empty means all services are accepted.
	Services []string
}

// Filter wraps an Entry channel and returns a new channel that only
// forwards entries matching the given options.
func Filter(in <-chan Entry, opts FilterOptions, stop <-chan struct{}) <-chan Entry {
	out := make(chan Entry, 64)
	levelSet := toSet(opts.Levels)
	svcSet := toSet(opts.Services)

	go func() {
		defer close(out)
		for {
			select {
			case entry, ok := <-in:
				if !ok {
					return
				}
				if !matchesService(entry.Service, svcSet) {
					continue
				}
				if !matchesLevel(entry.Line, levelSet) {
					continue
				}
				select {
				case out <- entry:
				case <-stop:
					return
				}
			case <-stop:
				return
			}
		}
	}()
	return out
}

func toSet(items []string) map[string]bool {
	if len(items) == 0 {
		return nil
	}
	s := make(map[string]bool, len(items))
	for _, v := range items {
		s[strings.ToLower(v)] = true
	}
	return s
}

func matchesService(svc string, set map[string]bool) bool {
	if set == nil {
		return true
	}
	return set[strings.ToLower(svc)]
}

func matchesLevel(line string, set map[string]bool) bool {
	if set == nil {
		return true
	}
	l := strings.ToLower(line)
	for level := range set {
		if strings.Contains(l, `"level":"`+level+`"`) ||
			strings.Contains(l, `"level": "`+level+`"`) {
			return true
		}
	}
	return false
}
