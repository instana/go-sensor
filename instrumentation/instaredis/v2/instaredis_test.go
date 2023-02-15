// (c) Copyright IBM Corp. 2023

//go:build go1.18
// +build go1.18

package instaredis_test

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"
	"github.com/instana/go-sensor/instrumentation/instaredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockPipeline struct {
	cmds    []redis.Cmder
	hooks   []redis.Hook
	current hooks
}

func (p *mockPipeline) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	cmd := redis.NewStatusCmd(ctx, []interface{}{"set", key, value}...)
	p.cmds = append(p.cmds, cmd)
	return cmd
}

func (p *mockPipeline) Incr(ctx context.Context, key string) *redis.IntCmd {
	cmd := redis.NewIntCmd(ctx, []interface{}{"incr", key}...)
	p.cmds = append(p.cmds, cmd)
	return cmd
}

func (p *mockPipeline) Do(ctx context.Context, args ...interface{}) *redis.Cmd {
	cmd := redis.NewCmd(ctx, args...)
	p.cmds = append(p.cmds, cmd)
	return cmd
}

func (p *mockPipeline) Exec(ctx context.Context) {
	p.process(ctx, p.cmds)
}

func (p *mockPipeline) process(ctx context.Context, cmds []redis.Cmder) (net.Conn, error) {
	switch ctx.Value("pipe_type") {
	case "processPipelineHook":
		return nil, p.current.pipeline(ctx, cmds)
	case "processTxPipelineHook":
		return nil, p.current.txPipeline(ctx, cmds)
	}
	return nil, errors.New("unidentified pipe type")
}

type mockClient struct {
	hooks   []redis.Hook
	options *redis.Options
	current hooks
}

type hooks struct {
	dial       redis.DialHook
	process    redis.ProcessHook
	pipeline   redis.ProcessPipelineHook
	txPipeline redis.ProcessPipelineHook
}

func newHooks() hooks {
	return hooks{
		dial:       func(ctx context.Context, network, addr string) (net.Conn, error) { return nil, nil },
		process:    func(ctx context.Context, cmd redis.Cmder) error { return nil },
		pipeline:   func(ctx context.Context, cmds []redis.Cmder) error { return nil },
		txPipeline: func(ctx context.Context, cmds []redis.Cmder) error { return nil },
	}
}

func (c *mockClient) chain() {
	for i := len(c.hooks) - 1; i >= 0; i-- {
		if wrapped := c.hooks[i].DialHook(c.current.dial); wrapped != nil {
			c.current.dial = wrapped
		}
		if wrapped := c.hooks[i].ProcessHook(c.current.process); wrapped != nil {
			c.current.process = wrapped
		}
		if wrapped := c.hooks[i].ProcessPipelineHook(c.current.pipeline); wrapped != nil {
			c.current.pipeline = wrapped
		}
		if wrapped := c.hooks[i].ProcessPipelineHook(c.current.txPipeline); wrapped != nil {
			c.current.txPipeline = wrapped
		}
	}
}

func newMockClient(options *redis.Options, foOptions *redis.FailoverOptions) *mockClient {
	if options != nil {
		return &mockClient{
			options: options,
			current: newHooks(),
		}
	}

	return &mockClient{
		options: &redis.Options{
			Addr: "FailoverClient",
			Dialer: func(ctx context.Context, network string, addr string) (net.Conn, error) {
				netConn := getMockConn(Single, ctx, network, addr)
				return netConn, nil
			},
		},
		current: newHooks(),
	}
}

func (c *mockClient) Close() error {
	return nil
}

func (c *mockClient) AddHook(hook redis.Hook) {
	c.hooks = append(c.hooks, hook)
	c.chain()
}

func (c mockClient) Options() *redis.Options {
	return c.options
}

func (c mockClient) Do(ctx context.Context, args ...interface{}) *redis.Cmd {
	cmd := redis.NewCmd(ctx, args...)
	c.runHooks(ctx, cmd)
	return cmd
}

func (c mockClient) Get(ctx context.Context, key string) *redis.StringCmd {
	cmd := redis.NewStringCmd(ctx, []interface{}{"get", key}...)
	c.runHooks(ctx, cmd)
	return cmd
}

func (c mockClient) Pipeline() mockPipeline {
	return mockPipeline{
		hooks:   c.hooks,
		current: c.current,
	}
}

func (c mockClient) TxPipeline() mockPipeline {
	return mockPipeline{
		hooks:   c.hooks,
		current: c.current,
	}
}

func (c mockClient) runHooks(ctx context.Context, cmd redis.Cmder) {
	c.current.process(ctx, cmd)
}

type mockClusterClient struct {
	hooks   []redis.Hook
	options *redis.ClusterOptions
	current hooks
}

func newMockClusterClient(options *redis.ClusterOptions, foOptions *redis.FailoverOptions) *mockClusterClient {
	if options != nil {
		return &mockClusterClient{
			options: options,
			current: newHooks(),
		}
	}

	return &mockClusterClient{
		options: &redis.ClusterOptions{
			Dialer: foOptions.Dialer,
		},
		current: newHooks(),
	}
}

func (c *mockClusterClient) Close() error {
	return nil
}

func (c *mockClusterClient) AddHook(hook redis.Hook) {
	c.hooks = append(c.hooks, hook)
	c.chain()
}

func (c mockClusterClient) Options() *redis.ClusterOptions {
	return c.options
}

func (c mockClusterClient) Do(ctx context.Context, args ...interface{}) *redis.Cmd {
	cmd := redis.NewCmd(ctx, args...)
	c.runHooks(ctx, cmd)
	return cmd
}

func (c mockClusterClient) Get(ctx context.Context, key string) *redis.StringCmd {
	cmd := redis.NewStringCmd(ctx, []interface{}{"get", key}...)
	c.runHooks(ctx, cmd)
	return cmd
}

func (c mockClusterClient) Pipeline() mockPipeline {
	return mockPipeline{
		hooks:   c.hooks,
		current: c.current,
	}
}

func (c mockClusterClient) TxPipeline() mockPipeline {
	return mockPipeline{
		hooks:   c.hooks,
		current: c.current,
	}
}

func (c *mockClusterClient) chain() {
	for i := len(c.hooks) - 1; i >= 0; i-- {
		if wrapped := c.hooks[i].DialHook(c.current.dial); wrapped != nil {
			c.current.dial = wrapped
		}
		if wrapped := c.hooks[i].ProcessHook(c.current.process); wrapped != nil {
			c.current.process = wrapped
		}
		if wrapped := c.hooks[i].ProcessPipelineHook(c.current.pipeline); wrapped != nil {
			c.current.pipeline = wrapped
		}
		if wrapped := c.hooks[i].ProcessPipelineHook(c.current.txPipeline); wrapped != nil {
			c.current.txPipeline = wrapped
		}
	}
}

func (c mockClusterClient) runHooks(ctx context.Context, cmd redis.Cmder) {
	c.current.process(ctx, cmd)
}

type redisType int

const (
	Single redisType = iota
	SingleFailover
	Cluster
	ClusterFailover
)

var redisTypeMap = map[redisType]string{
	Single:          "Single Server",
	SingleFailover:  "Single Failover Server",
	Cluster:         "Cluster Server",
	ClusterFailover: "Cluster Failover Server",
}

func buildNewClient(hasSentinel bool) *mockClient {
	if hasSentinel {
		return newMockClient(nil, &redis.FailoverOptions{
			RouteRandomly: false,
			MasterName:    "redis1",
			MaxRetries:    1,
			SentinelAddrs: []string{":26379"},
			Dialer: func(ctx context.Context, network string, addr string) (net.Conn, error) {
				netConn := getMockConn(SingleFailover, ctx, network, addr)
				return netConn, nil
			},
		})
	}

	return newMockClient(&redis.Options{
		Addr: ":6382",
		Dialer: func(ctx context.Context, network string, addr string) (net.Conn, error) {
			netConn := getMockConn(Single, ctx, network, addr)
			return netConn, nil
		},
	}, nil)
}

func buildNewClusterClient(hasSentinel bool) *mockClusterClient {
	if hasSentinel {
		return newMockClusterClient(nil, &redis.FailoverOptions{
			MasterName:    "redis1",
			MaxRetries:    1,
			SentinelAddrs: []string{":26379"},
			Dialer: func(ctx context.Context, network string, addr string) (net.Conn, error) {
				netConn := getMockConn(ClusterFailover, ctx, network, addr)
				return netConn, nil
			},
		})
	}

	return newMockClusterClient(&redis.ClusterOptions{
		Addrs: []string{":6382"},
		Dialer: func(ctx context.Context, network string, addr string) (net.Conn, error) {
			netConn := getMockConn(Cluster, ctx, network, addr)
			return netConn, nil
		},
	}, nil)
}

type MockAddr struct {
	network string
	addr    string
}

func (a MockAddr) Network() string {
	return a.network
}

func (a MockAddr) String() string {
	return a.addr
}

type MockConn struct {
	Ctx        context.Context
	Network    string
	Addr       string
	connData   []byte
	clientType redisType
}

func getMockConn(clientType redisType, ctx context.Context, network, addr string) *MockConn {
	conn := &MockConn{Ctx: ctx, Network: network, Addr: addr, clientType: clientType}
	return conn
}

func (c *MockConn) Read(b []byte) (n int, err error) {
	return 0, nil
}

func (c *MockConn) Write(b []byte) (n int, err error) {
	c.connData = append(c.connData, b...)
	return len(b), nil
}
func (c *MockConn) Close() error {
	return nil
}
func (c *MockConn) LocalAddr() net.Addr {
	return &MockAddr{c.Network, c.Addr}
}
func (c *MockConn) RemoteAddr() net.Addr {
	return &MockAddr{c.Network, c.Addr}
}
func (c *MockConn) SetDeadline(t time.Time) error {
	return nil
}
func (c *MockConn) SetReadDeadline(t time.Time) error {
	return nil
}
func (c *MockConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func TestClient(t *testing.T) {

	examples := map[string]struct {
		DoCommand       []interface{}
		DoPipeCommand   [][]interface{}
		DoTxPipeCommand [][]interface{}
		Expected        instana.RedisSpanTags
	}{
		"set name": {
			DoCommand: []interface{}{"set", "name", "Instana"},
			Expected: instana.RedisSpanTags{
				Command: "set",
			},
		},
		"get name": {
			DoCommand: []interface{}{"get", "name"},
			Expected: instana.RedisSpanTags{
				Command: "get",
			},
		},
		"del name": {
			DoCommand: []interface{}{"del", "name"},
			Expected: instana.RedisSpanTags{
				Command: "del",
			},
		},
		"batch commands with pipe": {
			DoPipeCommand: [][]interface{}{
				{"set", "name", "IBM"},
				{"get", "name"},
				{"del", "name"},
			},
			Expected: instana.RedisSpanTags{
				Command:     "multi",
				Subcommands: []string{"set", "get", "del"},
			},
		},
		"batch commands with tx pipe": {
			DoTxPipeCommand: [][]interface{}{
				{"set", "name", "IBM"},
				{"get", "name"},
				{"del", "name"},
			},
			Expected: instana.RedisSpanTags{
				Command:     "multi",
				Subcommands: []string{"set", "get", "del"},
			},
		},
	}

	rTypes := []redisType{Single, SingleFailover, Cluster, ClusterFailover}

	for _, rType := range rTypes {
		for name, example := range examples {
			t.Run(redisTypeMap[rType]+" - "+name, func(t *testing.T) {
				recorder := instana.NewTestRecorder()
				sensor := instana.NewSensorWithTracer(
					instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
				)

				sp := sensor.Tracer().StartSpan("testing")
				ctx := instana.ContextWithSpan(context.Background(), sp)

				if rType == Cluster || rType == ClusterFailover {
					rdb := buildNewClusterClient(rType == ClusterFailover)
					instaredis.WrapClusterClient(rdb, sensor)

					if len(example.DoCommand) > 0 {
						rdb.Do(ctx, example.DoCommand...)
					} else if len(example.DoPipeCommand) > 0 {
						pipe := rdb.Pipeline()

						for _, doCmd := range example.DoPipeCommand {
							pipe.Do(ctx, doCmd...)
						}
						ctx = context.WithValue(ctx, "pipe_type", "processPipelineHook")
						pipe.Exec(ctx)
					} else if len(example.DoTxPipeCommand) > 0 {
						pipe := rdb.TxPipeline()

						for _, doCmd := range example.DoTxPipeCommand {
							pipe.Do(ctx, doCmd...)
						}
						ctx = context.WithValue(ctx, "pipe_type", "processTxPipelineHook")
						pipe.Exec(ctx)
					}
					rdb.Close()
				} else {
					rdb := buildNewClient(rType == SingleFailover)
					instaredis.WrapClient(rdb, sensor)

					if len(example.DoCommand) > 0 {
						rdb.Do(ctx, example.DoCommand...)
					} else if len(example.DoPipeCommand) > 0 {
						pipe := rdb.Pipeline()

						for _, doCmd := range example.DoPipeCommand {
							pipe.Do(ctx, doCmd...)
						}
						ctx = context.WithValue(ctx, "pipe_type", "processPipelineHook")
						pipe.Exec(ctx)
					} else if len(example.DoTxPipeCommand) > 0 {
						pipe := rdb.TxPipeline()

						for _, doCmd := range example.DoTxPipeCommand {
							pipe.Do(ctx, doCmd...)
						}
						ctx = context.WithValue(ctx, "pipe_type", "processTxPipelineHook")
						pipe.Exec(ctx)
					}
					rdb.Close()
				}

				sp.Finish()

				spans := recorder.GetQueuedSpans()
				require.Len(t, spans, 2)

				dbSpan, parentSpan := spans[0], spans[1]

				assert.Equal(t, parentSpan.TraceID, dbSpan.TraceID)
				assert.Equal(t, parentSpan.TraceIDHi, dbSpan.TraceIDHi)
				assert.Equal(t, parentSpan.SpanID, dbSpan.ParentID)

				assert.Equal(t, "redis", dbSpan.Name)
				assert.EqualValues(t, instana.ExitSpanKind, dbSpan.Kind)
				assert.Empty(t, dbSpan.Ec)

				require.IsType(t, instana.RedisSpanData{}, dbSpan.Data)

				data := dbSpan.Data.(instana.RedisSpanData)

				assert.Equal(t, example.Expected.Error, data.Tags.Error)
				assert.Equal(t, example.Expected.Command, data.Tags.Command)

				if len(example.Expected.Subcommands) > 0 {
					assert.Equal(t, example.Expected.Subcommands, data.Tags.Subcommands)
				}
			})
		}
	}
}

type alwaysReadyClient struct{}

func (alwaysReadyClient) Ready() bool                                       { return true }
func (alwaysReadyClient) SendMetrics(data acceptor.Metrics) error           { return nil }
func (alwaysReadyClient) SendEvent(event *instana.EventData) error          { return nil }
func (alwaysReadyClient) SendSpans(spans []instana.Span) error              { return nil }
func (alwaysReadyClient) SendProfiles(profiles []autoprofile.Profile) error { return nil }
func (alwaysReadyClient) Flush(context.Context) error                       { return nil }
