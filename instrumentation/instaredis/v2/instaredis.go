// (c) Copyright IBM Corp. 2023

//go:build go1.18
// +build go1.18

package instaredis

import (
	"context"
	"net"
	"strings"

	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/redis/go-redis/v9"
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

type InstanaRedisClient interface {
	AddHook(hook redis.Hook)
	Options() *redis.Options
}

// DialHook adds a middleware to the existing DialHook. This is required to satisfy Hook interface.
func (h commandCaptureHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		conn, err := next(ctx, network, addr)
		return conn, err
	}
}

// ProcessHook adds an instrumentation middleware to the existing ProcessHook
func (h commandCaptureHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		h.handleHook(ctx, cmd, []redis.Cmder{})
		err := next(ctx, cmd)
		return err
	}
}

// ProcessPipelineHook adds an instrumentation middleware to the existing ProcessPipelineHook
func (h commandCaptureHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		h.handleHook(ctx, nil, cmds)
		err := next(ctx, cmds)
		return err
	}
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
