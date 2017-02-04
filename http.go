package instana

const (
	HTTPClient = "g.hc"
	HTTPServer = "g.http"
)

type HTTPData struct {
	Host   string `json:"host"`
	URL    string `json:"url"`
	Status int    `json:"status"`
	Method string `json:"method"`
}
