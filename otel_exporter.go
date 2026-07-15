package instana

import (
	"context"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type instanaExporter struct{}

// ExportSpans is called by OpenTelemetry when spans are ready to be exported.
func (e *instanaExporter) ExportSpans(
	ctx context.Context,
	spans []sdktrace.ReadOnlySpan,
) error {

	// Future work:
	// - convert spans and send them to the Instana agent

	//loop through the exported spans
	for _, span := range spans {
		_ = span.Name() //access to the span namee
	}

	return nil
}

// Shutdown is called when the exporter is shutting down, means nothing to clean up for this POC
func (e *instanaExporter) Shutdown(
	ctx context.Context,
) error {
	return nil
}
