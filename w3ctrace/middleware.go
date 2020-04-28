package w3ctrace

import "net/http"

// TracingHandlerFunc is an HTTP middleware that forwards the W3C context found in request
// with the response
func TracingHandlerFunc(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if trCtx, err := Extract(req.Header); err == nil {
			Inject(trCtx, w.Header())
		}

		handler(w, req)
	}
}
