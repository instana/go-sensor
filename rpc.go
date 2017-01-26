package instana

const (
	RPC = "g.rpc"
)

type RPCData struct {
	Host string `json:"host"`
	Call string `json:"call"`
}
