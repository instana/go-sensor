// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

// Package instaawssdk instruments github.com/aws/aws-sdk-go

package instaawssdk

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	instana "github.com/instana/go-sensor"
	otlog "github.com/opentracing/opentracing-go/log"
)

var errMethodNotInstrumented = errors.New("method not instrumented")

// InstrumentSession instruments github.com/aws/aws-sdk-go/aws/session.Session by
// injecting handlers to create and finalize Instana spans while sending and completing
// requests
func InstrumentSession(sess *session.Session, sensor *instana.Sensor) {
	sess.Handlers.Send.PushFront(func(req *request.Request) {
		switch req.ClientInfo.ServiceName {
		case s3.ServiceName:
			StartS3Span(req, sensor)
		}
	})

	sess.Handlers.Complete.PushBack(func(req *request.Request) {
		sp, ok := instana.SpanFromContext(req.Context())
		if !ok {
			return
		}
		defer sp.Finish()

		if req.Error != nil {
			sp.LogFields(otlog.Error(req.Error))
		}

		switch req.ClientInfo.ServiceName {
		case s3.ServiceName:
			FinalizeS3Span(sp, req)
		}
	})
}
