package instana

type Data struct {
	Service string   `json:"service,omitempty"`
	SDK     *SDKData `json:"sdk"`
}
