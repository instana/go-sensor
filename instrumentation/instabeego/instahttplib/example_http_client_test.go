package instahttplib_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/beego/beego/v2/client/httplib"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instabeego/instahttplib"
)

func main(t *testing.T) {
	sensor := instana.NewSensor("my-http-client")
	span := sensor.Tracer().StartSpan("entry")

	defer span.Finish()

	ctx := instana.ContextWithSpan(context.Background(), span)

	instahttplib.Instrument(sensor)

	req := httplib.NewBeegoRequestWithCtx(ctx, "https://www.instana.com", http.MethodGet)
	req.DoRequest()
}
