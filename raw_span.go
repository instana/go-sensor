package instana

type jsonSpan struct {
	TraceID   int64  `json:"t"`
	ParentID  *int64 `json:"p,omitempty"`
	SpanID    int64  `json:"s"`
	Timestamp uint64 `json:"ts"`
	Duration  uint64 `json:"d"`
	Name      string `json:"n"`
	From      *FromS `json:"f"`
	Data      *Data  `json:"data"`
}
