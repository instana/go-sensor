package instasarama

import (
	"github.com/Shopify/sarama"
	ot "github.com/opentracing/opentracing-go"
)

// ProducerMessageWithSpan injects the tracing context into producer message headers to propagate
// them through the Kafka requests made with instasarama producers.
func ProducerMessageWithSpan(pm *sarama.ProducerMessage, sp ot.Span) *sarama.ProducerMessage {
	sp.Tracer().Inject(sp.Context(), ot.TextMap, ProducerMessageCarrier{Message: pm})
	return pm
}
