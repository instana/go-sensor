// (c) Copyright IBM Corp. 2022

//go:build go1.13
// +build go1.13

package instaredis

import (
	"context"
	"strings"

	"github.com/go-redis/redis/v8"
	instana "github.com/instana/go-sensor"
	"github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"
)

type commandCaptureHook struct {
	options *redis.Options
	sensor  *instana.Sensor
}

func setSpanCommands(span opentracing.Span, cmd redis.Cmder, cmds []redis.Cmder) {
	if cmd != nil {
		span.SetTag("redis.command", cmd.Name())

		if cmd.Err() != nil {
			span.SetTag("redis.error", cmd.Err().Error())
			span.LogFields(otlog.Object("error", cmd.Err().Error()))
		}
	} else if len(cmds) > 0 {
		var subCommands []string

		if cmds[0].Name() == "multi" && cmds[len(cmds)-1].Name() == "exec" {
			for i := 1; i < len(cmds)-1; i++ {
				subCommands = append(subCommands, cmds[i].Name())
			}
		} else {
			for _, cmd := range cmds {
				subCommands = append(subCommands, cmd.Name())
			}
		}
		span.SetTag("redis.command", "multi")
		span.SetTag("redis.subCommands", subCommands)
	}
}

func (h commandCaptureHook) handleHook(ctx context.Context, cmd redis.Cmder, cmds []redis.Cmder) {

	connection := h.options.Addr

	if connection == "FailoverClient" {
		conn, err := h.options.Dialer(ctx, h.options.Network, "")

		if err == nil {
			connection = conn.RemoteAddr().String()
		}
	}

	// if the IP provided in the Redis constructor have only ports. eg: :6379, :6380 and so on, we add 127.0.0.1
	if strings.HasPrefix(connection, ":") {
		connection = "127.0.0.1" + connection
	}

	tags := opentracing.Tags{
		"redis.connection": connection,
	}

	var span opentracing.Span

	// We need an entry parent span in order to test this, so we will need a webserver or manually create an entry span
	parentSpan, ok := instana.SpanFromContext(ctx)

	if ok {
		span = h.sensor.Tracer().StartSpan("redis", opentracing.ChildOf(parentSpan.Context()), tags)
	} else {
		span = h.sensor.Tracer().StartSpan("redis", tags)
	}

	setSpanCommands(span, cmd, cmds)

	defer span.Finish()
}

func (h commandCaptureHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	return ctx, nil
}

func (h commandCaptureHook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	h.handleHook(ctx, cmd, nil)
	return nil
}

func (h commandCaptureHook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	return ctx, nil
}

func (h commandCaptureHook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	h.handleHook(ctx, nil, cmds)
	return nil
}

// WrapClient TODO: rename it to something more meaningful
// WrapClient wraps the Redis client instance in order to add the instrumentation
func WrapClient(client *redis.Client, sensor *instana.Sensor) *redis.Client {
	opts := client.Options()
	client.AddHook(&commandCaptureHook{opts, sensor})
	return client
}

type clusterCommandCaptureHook struct {
	options *redis.ClusterOptions
	sensor  *instana.Sensor
}

func (h clusterCommandCaptureHook) handleHook(ctx context.Context, cmd redis.Cmder, cmds []redis.Cmder) {

	var connection string

	if len(h.options.Addrs) > 0 {
		connection = h.options.Addrs[0]
	} else if h.options.ClusterSlots != nil {
		cs, err := h.options.ClusterSlots(ctx)

		if err == nil {
			for _, clusterSlot := range cs {
				connection = clusterSlot.Nodes[0].Addr
			}
		}
	}

	// if the IP provided in the Redis constructor have only ports. eg: :6379, :6380 and so on, we add 127.0.0.1
	if strings.HasPrefix(connection, ":") {
		connection = "127.0.0.1" + connection
	}

	tags := opentracing.Tags{
		"redis.connection": connection,
	}

	var span opentracing.Span

	// We need an entry parent span in order to test this, so we will need a webserver or manually create an entry span
	parentSpan, ok := instana.SpanFromContext(ctx)

	if ok {
		span = h.sensor.Tracer().StartSpan("redis", opentracing.ChildOf(parentSpan.Context()), tags)
	} else {
		span = h.sensor.Tracer().StartSpan("redis", tags)
	}

	setSpanCommands(span, cmd, cmds)

	defer span.Finish()
}

func (h clusterCommandCaptureHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	return ctx, nil
}

func (h clusterCommandCaptureHook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	h.handleHook(ctx, cmd, nil)
	return nil
}

func (h clusterCommandCaptureHook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	return ctx, nil
}

func (h clusterCommandCaptureHook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	h.handleHook(ctx, nil, cmds)
	return nil
}

// WrapClient wraps the Redis client instance in order to add the instrumentation
func WrapClusterClient(clusterClient *redis.ClusterClient, sensor *instana.Sensor) *redis.ClusterClient {
	opts := clusterClient.Options()
	clusterClient.AddHook(&clusterCommandCaptureHook{opts, sensor})
	return clusterClient
}
