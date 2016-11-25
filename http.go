package instana

const (
	HTTP_CLIENT = "g.hc"
	HTTP_SERVER = "g.http"
)

type HttpData struct {
	Host   string `json:"host"`
	Url    string `json:"url"`
	Status int    `json:"status"`
	Method string `json:"method"`
}
