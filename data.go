package instana

type Data struct {
	Service string            `json:"service"`
	SDK     *SDKData          `json:"sdk,omitempty"`
	HTTP    *HTTPData         `json:"http,omitempty"`
	RPC     *RPCData          `json:"rpc,omitempty"`
	Baggage map[string]string `json:"baggage,omitempty"`
	Custom  *CustomData       `json:"custom,omitempty"`
}
