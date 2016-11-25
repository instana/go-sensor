package instana

const (
	RPC = "g.rpc"
)

type RpcData struct {
	Host string `json:"host"`
	Call string `json:"call"`
}
