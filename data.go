package instana

type Data struct {
	Service string            `json:"service"`
	Http    *HttpData         `json:"http,omitempty"`
	Rpc     *RpcData          `json:"rpc,omitempty"`
	Baggage map[string]string `json:"baggage,omitempty"`
}
