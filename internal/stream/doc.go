// Package stream provides fan-in multiplexing and filtering of structured
// log entries across multiple service sources.
//
// Usage:
//
//	sources := map[string]string{
//		"api":    "/var/log/api.log",
//		"worker": "/var/log/worker.log",
//	}
//
//	m, err := stream.New(sources)
//	if err != nil {
//		log.Fatal(err)
//	}
//	m.Start()
//	defer m.Stop()
//
//	stop := make(chan struct{})
//	filtered := stream.Filter(m.Out(), stream.FilterOptions{
//		Levels:   []string{"error", "warn"},
//		Services: []string{"api"},
//	}, stop)
//
//	for entry := range filtered {
//		fmt.Printf("[%s] %s\n", entry.Service, entry.Line)
//	}
package stream
