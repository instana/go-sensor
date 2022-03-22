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
	options        *redis.Options
	clusterOptions *redis.ClusterOptions
	sensor         *instana.Sensor
}

func setSpanCommands(span opentracing.Span, cmd redis.Cmder, cmds []redis.Cmder) {
	if cmd != nil {
		span.SetTag("redis.command", cmd.Name())

		if cmd.Err() != nil {
			span.SetTag("redis.error", cmd.Err().Error())
			span.LogFields(otlog.Object("error", cmd.Err().Error()))
		}
		return
	}

	i, end := 0, len(cmds)

	if cmds[0].Name() == "multi" && cmds[len(cmds)-1].Name() == "exec" {
		i = 1
		end = end - 1
	}

	var subCommands []string

	for ; i < end; i++ {
		subCommands = append(subCommands, cmds[i].Name())

		if cmds[i].Err() != nil {
			span.SetTag("redis.error", cmds[i].Err().Error())
			span.LogFields(otlog.Object("error", cmds[i].Err().Error()))
		}
	}

	span.SetTag("redis.command", "multi")
	span.SetTag("redis.subCommands", subCommands)
}

func (h commandCaptureHook) getConnection(ctx context.Context) string {
	if h.clusterOptions != nil {
		if len(h.clusterOptions.Addrs) > 0 {
			return h.clusterOptions.Addrs[0]
		}

		if h.clusterOptions.ClusterSlots != nil {
			cs, err := h.clusterOptions.ClusterSlots(ctx)

			if err == nil {
				return cs[0].Nodes[0].Addr
			}
		}
		return ""
	}

	connection := h.options.Addr

	if connection == "FailoverClient" {
		conn, err := h.options.Dialer(ctx, h.options.Network, "")

		if err == nil {
			return conn.RemoteAddr().String()
		}
	}
	return connection
}

func (h commandCaptureHook) handleHook(ctx context.Context, cmd redis.Cmder, cmds []redis.Cmder) {
	connection := h.getConnection(ctx)

	// if the IP provided in the Redis constructor have only ports. eg: :6379, :6380 and so on, we add localhost
	if strings.HasPrefix(connection, ":") {
		connection = "localhost" + connection
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

	span.Finish()
}

func (h commandCaptureHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	return ctx, nil
}

func (h commandCaptureHook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	h.handleHook(ctx, cmd, []redis.Cmder{})
	return nil
}

func (h commandCaptureHook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	return ctx, nil
}

func (h commandCaptureHook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	h.handleHook(ctx, nil, cmds)
	return nil
}

// WrapClient wraps the Redis client instance in order to add the instrumentation
func WrapClient(client *redis.Client, sensor *instana.Sensor) *redis.Client {
	opts := client.Options()
	client.AddHook(&commandCaptureHook{options: opts, sensor: sensor})
	return client
}

// WrapClient wraps the Redis client instance in order to add the instrumentation
func WrapClusterClient(clusterClient *redis.ClusterClient, sensor *instana.Sensor) *redis.ClusterClient {
	opts := clusterClient.Options()
	clusterClient.AddHook(&commandCaptureHook{clusterOptions: opts, sensor: sensor})
	return clusterClient
}
