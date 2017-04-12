package instana

type SDKData struct {
	Name      string      `json:"name"`
	Type      string      `json:"type"`
	Arguments string      `json:"arguments"`
	Return    string      `json:"return"`
	Custom    *CustomData `json:"custom,omitempty"`
}
