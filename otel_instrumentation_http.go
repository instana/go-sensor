package instana

import (
	"net/http"

	"go.opentelemetry.io/otel"
)

//Basic HTTP middleware using opentelemetry

func OTelTracingHandlerFunc(
	pathTemplate string,
	handler http.HandlerFunc,
) http.HandlerFunc {

	return func(
		w http.ResponseWriter,
		r *http.Request,
	) {

		tracer := otel.Tracer("instana") //creating a tracer for incoming requests

		ctx, span := tracer.Start( //start anew span for this request
			r.Context(),
			"g.http",
		)
		defer span.End() //end the span when the request finishes

		handler( //pass the updated context to the handler
			w,
			r.WithContext(ctx),
		)
	}
}
