package instana

type Data struct {
	Service string   `json:"service"`
	SDK     *SDKData `json:"sdk,omitempty"`
}
