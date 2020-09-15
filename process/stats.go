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

// LimitedResource represents a limited resource stat
type LimitedResource struct {
	Current int
	Max     int
}

// ResourceLimits represents resource limits for a process
type ResourceLimits struct {
	OpenFiles LimitedResource
}
