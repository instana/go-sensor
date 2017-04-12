package instana

type SDKData struct {
	Name      string      `json:"name"`
	Type      string      `json:"type,omitempty"`
	Arguments string      `json:"arguments,omitempty"`
	Return    string      `json:"return,omitempty"`
	Custom    *CustomData `json:"custom,omitempty"`
}
