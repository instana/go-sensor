package instana

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

//check that the propagator exposes the expected requests

func TestInstanaPropagatorFields(t *testing.T) {
	p := instanaPropagator{}

	fields := p.Fields() //get the headers supported by the propagator

	//make sure all the expected headers are returned
	if len(fields) != 3 {
		t.Fatalf("expected 3 fields, got %d", len(fields))
	}
}

//check that the trace information can be rebuilt from headers

func TestInstanaPropagatorExtract(t *testing.T) {
	p := instanaPropagator{}

	// use instana sample headers
	carrier := propagation.MapCarrier{
		"x-instana-t": "1234567890abcdef1234567890abcdef",
		"x-instana-s": "1234567890abcdef",
		"x-instana-l": "1",
	}

	ctx := p.Extract(context.Background(), carrier) //extract trace information from the headers

	sc := trace.SpanContextFromContext(ctx) //read the span context

	//make sure that a valid span context is made
	if !sc.IsValid() {
		t.Fatal("expected valid span context")
	}
}
