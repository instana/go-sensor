// (c) Copyright IBM Corp. 2025

package instana

type spanCategory string

const (
	logging spanCategory = "logging"
	unknown spanCategory = "unknown"
)

func (c spanCategory) String() string {
	return string(c)
}

func (opts *TracerOptions) DisableAllCategories() {
	opts.Disable = map[string]bool{
		logging.String(): true,
	}
}

// registeredSpanMap maps span types to their categories
var registeredSpanMap map[RegisteredSpanType]spanCategory = map[RegisteredSpanType]spanCategory{
	// logging
	LogSpanType: logging,
}

func (r *spanS) getSpanCategory() spanCategory {
	// return span category if it is a registered span type
	if c, ok := registeredSpanMap[RegisteredSpanType(r.Operation)]; ok {
		return c
	}

	return unknown
}

func (c spanCategory) Enabled() bool {
	return !sensor.options.Tracer.Disable[c.String()]
}
