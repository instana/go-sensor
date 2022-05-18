// (c) Copyright IBM Corp. 2022

//go:build go1.13
// +build go1.13

package instaredis

import (
	"context"
	"strings"

	"github.com/go-redis/redis/v8"
	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"
)

type commandCaptureHook struct {
	options        *redis.Options
	clusterOptions *redis.ClusterOptions
	sensor         *instana.Sensor
	connection     string
}

func newCommandCapture(s instana.Sensor, o *redis.Options, co *redis.ClusterOptions) *commandCaptureHook {
	var cch *commandCaptureHook

	if o != nil {
		cch = &commandCaptureHook{options: o, sensor: &s, connection: ""}
		cch.connection = cch.options.Addr
	} else {
		cch = &commandCaptureHook{clusterOptions: co, sensor: &s, connection: ""}

		if cch.clusterOptions != nil && len(cch.clusterOptions.Addrs) > 0 {
			cch.connection = cch.clusterOptions.Addrs[0]
		}
	}

	return cch
}

func setSpanCommands(span ot.Span, cmd redis.Cmder, cmds []redis.Cmder) {
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
	if h.connection == "FailoverClient" {
		conn, err := h.options.Dialer(ctx, h.options.Network, "")

		if err == nil {
			h.connection = conn.RemoteAddr().String()
			return h.connection
		}
	}

	if h.connection != "" {
		return h.connection
	}

	if h.clusterOptions != nil && h.clusterOptions.ClusterSlots != nil {
		if cs, err := h.clusterOptions.ClusterSlots(ctx); err == nil {
			h.connection = cs[0].Nodes[0].Addr
			return h.connection
		}
		return ""
	}

	return ""
}

func (h commandCaptureHook) handleHook(ctx context.Context, cmd redis.Cmder, cmds []redis.Cmder) {
	connection := h.getConnection(ctx)

	// if the IP provided in the Redis constructor have only ports. eg: :6379, :6380 and so on, we add localhost
	if strings.HasPrefix(connection, ":") {
		connection = "localhost" + connection
	}

	opts := []ot.StartSpanOption{
		ot.Tags{
			"redis.connection": connection,
		},
	}

	// We need an entry parent span in order to test this, so we will need a webserver or manually create an entry span
	tracer := h.sensor.Tracer()
	var span ot.Span = tracer.StartSpan("redis", opts...)

	if ps, ok := instana.SpanFromContext(ctx); ok {
		tracer = ps.Tracer()
		opts = append(opts, ot.ChildOf(ps.Context()))
		span = tracer.StartSpan("redis", opts...)
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

type InstanaRedisClient interface {
	AddHook(hook redis.Hook)
	Options() *redis.Options
}

type InstanaRedisClusterClient interface {
	AddHook(hook redis.Hook)
	Options() *redis.ClusterOptions
}

// WrapClient wraps the Redis client instance in order to add the instrumentation
func WrapClient(client InstanaRedisClient, sensor *instana.Sensor) InstanaRedisClient {
	opts := client.Options()
	client.AddHook(newCommandCapture(*sensor, opts, nil))
	return client
}

// WrapClusterClient wraps the Redis client instance in order to add the instrumentation
func WrapClusterClient(clusterClient InstanaRedisClusterClient, sensor *instana.Sensor) InstanaRedisClusterClient {
	opts := clusterClient.Options()
	clusterClient.AddHook(newCommandCapture(*sensor, nil, opts))
	return clusterClient
}
