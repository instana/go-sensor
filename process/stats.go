package process

// MemStats represents memory stats for a process
type MemStats struct {
	Total  int
	Rss    int
	Shared int
}

// CPUStats represents CPU stats for a process
type CPUStats struct {
	User   int
	System int
}
