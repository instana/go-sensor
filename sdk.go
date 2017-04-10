package instana

type SDKData struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	SpanKind string `json:"span.kind"`
}
