// (c) Copyright IBM Corp. 2023
// (c) Copyright Instana Inc. 2023

//go:build go1.21
// +build go1.21

package instahttplib

import (
	"github.com/beego/beego/v2/client/httplib"
	instana "github.com/instana/go-sensor"
)

func Instrument(sensor *instana.Sensor) {
	httplib.WithTransport(instana.RoundTripper(sensor, nil))
}

func NewClient(sensor *instana.Sensor, name string, endpoint string,
	opts ...httplib.ClientOption) (*httplib.Client, error) {

	httplib.WithTransport(instana.RoundTripper(sensor, nil))
	return httplib.NewClient(name, endpoint, opts...)
}
