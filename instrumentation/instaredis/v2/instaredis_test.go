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

const testErrStr string = "Random test error!"

type mockPipeline struct {
	cmds    []redis.Cmder
	hooks   []redis.Hook
	current hooks

	isError bool
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
	if p.isError {
		p.cmds[0].SetErr(errors.New(testErrStr))
	}
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

	isError bool
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
		process:    func(ctx context.Context, cmd redis.Cmder) error { return cmd.Err() },
		pipeline:   func(ctx context.Context, cmds []redis.Cmder) error { return cmds[0].Err() },
		txPipeline: func(ctx context.Context, cmds []redis.Cmder) error { return cmds[0].Err() },
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

func newMockClient(options *redis.Options, foOptions *redis.FailoverOptions, isError bool) *mockClient {
	if options != nil {
		return &mockClient{
			options: options,
			current: newHooks(),

			isError: isError,
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
		isError: isError,
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
	if c.isError {
		cmd.SetErr(errors.New(testErrStr))
	}
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

		isError: c.isError,
	}
}

func (c mockClient) TxPipeline() mockPipeline {
	return mockPipeline{
		hooks:   c.hooks,
		current: c.current,

		isError: c.isError,
	}
}

func (c mockClient) runHooks(ctx context.Context, cmd redis.Cmder) {
	c.current.process(ctx, cmd)
}

type mockClusterClient struct {
	hooks   []redis.Hook
	options *redis.ClusterOptions
	current hooks

	isError bool
}

func newMockClusterClient(options *redis.ClusterOptions, foOptions *redis.FailoverOptions, isError bool) *mockClusterClient {
	if options != nil {
		return &mockClusterClient{
			options: options,
			current: newHooks(),
			isError: isError,
		}
	}

	return &mockClusterClient{
		options: &redis.ClusterOptions{
			Dialer: foOptions.Dialer,
		},
		current: newHooks(),
		isError: isError,
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
	if c.isError {
		cmd.SetErr(errors.New(testErrStr))
	}
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

		isError: c.isError,
	}
}

func (c mockClusterClient) TxPipeline() mockPipeline {
	return mockPipeline{
		hooks:   c.hooks,
		current: c.current,

		isError: c.isError,
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

func buildNewClient(hasSentinel bool, isError bool) *mockClient {
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
		}, isError)
	}

	return newMockClient(&redis.Options{
		Addr: ":6382",
		Dialer: func(ctx context.Context, network string, addr string) (net.Conn, error) {
			netConn := getMockConn(Single, ctx, network, addr)
			return netConn, nil
		},
	}, nil, isError)
}

func buildNewClusterClient(hasSentinel bool, isError bool) *mockClusterClient {
	if hasSentinel {
		return newMockClusterClient(nil, &redis.FailoverOptions{
			MasterName:    "redis1",
			MaxRetries:    1,
			SentinelAddrs: []string{":26379"},
			Dialer: func(ctx context.Context, network string, addr string) (net.Conn, error) {
				netConn := getMockConn(ClusterFailover, ctx, network, addr)
				return netConn, nil
			},
		}, isError)
	}

	return newMockClusterClient(&redis.ClusterOptions{
		Addrs: []string{":6382"},
		Dialer: func(ctx context.Context, network string, addr string) (net.Conn, error) {
			netConn := getMockConn(Cluster, ctx, network, addr)
			return netConn, nil
		},
	}, nil, isError)
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
		doCommand       []interface{}
		doPipeCommand   [][]interface{}
		doTxPipeCommand [][]interface{}
		expected        instana.RedisSpanTags

		isError bool
	}{
		"set name": {
			doCommand: []interface{}{"set", "name", "Instana"},
			expected: instana.RedisSpanTags{
				Command: "set",
			},
		},
		"set name error": {
			doCommand: []interface{}{"set", "name", "Instana"},
			expected: instana.RedisSpanTags{
				Command: "set",
				Error:   testErrStr,
			},

			isError: true,
		},
		"get name": {
			doCommand: []interface{}{"get", "name"},
			expected: instana.RedisSpanTags{
				Command: "get",
			},
		},
		"get name error": {
			doCommand: []interface{}{"get", "name"},
			expected: instana.RedisSpanTags{
				Command: "get",
				Error:   testErrStr,
			},
			isError: true,
		},
		"del name": {
			doCommand: []interface{}{"del", "name"},
			expected: instana.RedisSpanTags{
				Command: "del",
			},
		},
		"del name error": {
			doCommand: []interface{}{"del", "name"},
			expected: instana.RedisSpanTags{
				Command: "del",
				Error:   testErrStr,
			},
			isError: true,
		},
		"batch commands with pipe": {
			doPipeCommand: [][]interface{}{
				{"set", "name", "IBM"},
				{"get", "name"},
				{"del", "name"},
			},
			expected: instana.RedisSpanTags{
				Command:     "multi",
				Subcommands: []string{"set", "get", "del"},
			},
		},
		"batch commands with pipe error": {
			doPipeCommand: [][]interface{}{
				{"set", "name", "IBM"},
				{"get", "name"},
				{"del", "name"},
			},
			expected: instana.RedisSpanTags{
				Command:     "multi",
				Subcommands: []string{"set", "get", "del"},
				Error:       testErrStr,
			},
			isError: true,
		},
		"batch commands with tx pipe": {
			doTxPipeCommand: [][]interface{}{
				{"set", "name", "IBM"},
				{"get", "name"},
				{"del", "name"},
			},
			expected: instana.RedisSpanTags{
				Command:     "multi",
				Subcommands: []string{"set", "get", "del"},
			},
		},
		"batch commands with tx pipe error": {
			doTxPipeCommand: [][]interface{}{
				{"set", "name", "IBM"},
				{"get", "name"},
				{"del", "name"},
			},
			expected: instana.RedisSpanTags{
				Command:     "multi",
				Subcommands: []string{"set", "get", "del"},
				Error:       testErrStr,
			},
			isError: true,
		},
	}

	rTypes := []redisType{Single, SingleFailover, Cluster, ClusterFailover}

	for _, rType := range rTypes {
		for name, example := range examples {
			t.Run(redisTypeMap[rType]+" - "+name, func(t *testing.T) {
				recorder := instana.NewTestRecorder()
				c := instana.InitCollector(&instana.Options{
					AgentClient: alwaysReadyClient{},
					Recorder:    recorder,
				})
				defer instana.ShutdownCollector()

				sp := c.Tracer().StartSpan("testing")
				ctx := instana.ContextWithSpan(context.Background(), sp)

				if rType == Cluster || rType == ClusterFailover {
					rdb := buildNewClusterClient(rType == ClusterFailover, example.isError)
					instaredis.WrapClusterClient(rdb, c)

					if len(example.doCommand) > 0 {
						rdb.Do(ctx, example.doCommand...)
					} else if len(example.doPipeCommand) > 0 {
						pipe := rdb.Pipeline()

						for _, doCmd := range example.doPipeCommand {
							pipe.Do(ctx, doCmd...)
						}
						ctx = context.WithValue(ctx, "pipe_type", "processPipelineHook")
						pipe.Exec(ctx)
					} else if len(example.doTxPipeCommand) > 0 {
						pipe := rdb.TxPipeline()

						for _, doCmd := range example.doTxPipeCommand {
							pipe.Do(ctx, doCmd...)
						}
						ctx = context.WithValue(ctx, "pipe_type", "processTxPipelineHook")
						pipe.Exec(ctx)
					}
					rdb.Close()
				} else {
					rdb := buildNewClient(rType == SingleFailover, example.isError)
					instaredis.WrapClient(rdb, c)

					if len(example.doCommand) > 0 {
						rdb.Do(ctx, example.doCommand...)
					} else if len(example.doPipeCommand) > 0 {
						pipe := rdb.Pipeline()

						for _, doCmd := range example.doPipeCommand {
							pipe.Do(ctx, doCmd...)
						}
						ctx = context.WithValue(ctx, "pipe_type", "processPipelineHook")
						pipe.Exec(ctx)
					} else if len(example.doTxPipeCommand) > 0 {
						pipe := rdb.TxPipeline()

						for _, doCmd := range example.doTxPipeCommand {
							pipe.Do(ctx, doCmd...)
						}
						ctx = context.WithValue(ctx, "pipe_type", "processTxPipelineHook")
						pipe.Exec(ctx)
					}
					rdb.Close()
				}

				sp.Finish()

				spans := recorder.GetQueuedSpans()

				var dbSpan, parentSpan, logSpan instana.Span

				// if this is an error case
				if example.isError {
					require.Len(t, spans, 3)
					dbSpan, logSpan, parentSpan = spans[0], spans[1], spans[2]
					// assert error
					assert.Equal(t, parentSpan.TraceID, logSpan.TraceID)
					assert.Equal(t, dbSpan.SpanID, logSpan.ParentID)
					assert.Equal(t, "log.go", logSpan.Name)

					// assert that log message has been recorded within the span interval
					assert.GreaterOrEqual(t, logSpan.Timestamp, dbSpan.Timestamp)
					assert.LessOrEqual(t, logSpan.Duration, dbSpan.Duration)

					require.IsType(t, instana.LogSpanData{}, logSpan.Data)
					logData := logSpan.Data.(instana.LogSpanData)

					assert.Equal(t, instana.LogSpanTags{
						Level:   "ERROR",
						Message: `error: "Random test error!"`,
					}, logData.Tags)
					assert.Equal(t, 1, dbSpan.Ec)

				} else {
					require.Len(t, spans, 2)
					dbSpan, parentSpan = spans[0], spans[1]
					assert.Empty(t, dbSpan.Ec)
				}

				assert.Equal(t, parentSpan.TraceID, dbSpan.TraceID)
				assert.Equal(t, parentSpan.TraceIDHi, dbSpan.TraceIDHi)
				assert.Equal(t, parentSpan.SpanID, dbSpan.ParentID)

				assert.Equal(t, "redis", dbSpan.Name)
				assert.EqualValues(t, instana.ExitSpanKind, dbSpan.Kind)

				require.IsType(t, instana.RedisSpanData{}, dbSpan.Data)

				data := dbSpan.Data.(instana.RedisSpanData)

				assert.Equal(t, example.expected.Error, data.Tags.Error)
				assert.Equal(t, example.expected.Command, data.Tags.Command)

				if len(example.expected.Subcommands) > 0 {
					assert.Equal(t, example.expected.Subcommands, data.Tags.Subcommands)
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
