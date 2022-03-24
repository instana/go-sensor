// (c) Copyright IBM Corp. 2022

//go:build go1.13
// +build go1.13

package instaredis_test

import (
	"context"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instaredis"
	"github.com/instana/testify/assert"
	"github.com/instana/testify/require"
)

/*
Redis small cheat sheet (command sent vs characters received):

SET  = +OK\r\n
GET  = $4\r\n Where 4 is the size of the item "gotten"
DEL  = :1\r\n Where 1 is the amount of data deleted (I guess)
INCR = :7\r\n Where 7 is the new value of the variable incremented
*/

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

// returns an instance of redis.Client, compatible with redis.NewClient() and redis.NewFailoverClient()
func buildNewClient(hasSentinel bool) *redis.Client {
	if hasSentinel {
		return redis.NewFailoverClient(&redis.FailoverOptions{
			SlaveOnly:     true,
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

	return redis.NewClient(&redis.Options{
		Addr: ":6382",
		Dialer: func(ctx context.Context, network string, addr string) (net.Conn, error) {
			netConn := getMockConn(Single, ctx, network, addr)
			return netConn, nil
		},
	})
}

// returns an instance of redis.ClusterClient, compatible with redis.NewClusterClient() and redis.NewFailoverClusterClient()
func buildNewClusterClient(hasSentinel bool) *redis.ClusterClient {
	if hasSentinel {
		return redis.NewFailoverClusterClient(&redis.FailoverOptions{
			MasterName:    "redis1",
			MaxRetries:    1,
			SentinelAddrs: []string{":26379"},
			Dialer: func(ctx context.Context, network string, addr string) (net.Conn, error) {
				netConn := getMockConn(ClusterFailover, ctx, network, addr)
				return netConn, nil
			},
		})
	}

	return redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{":6382"},
		Dialer: func(ctx context.Context, network string, addr string) (net.Conn, error) {
			netConn := getMockConn(Cluster, ctx, network, addr)
			return netConn, nil
		},
	})
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

func parseIncomingCommand(incomingCmd []byte, isSingleRedis bool) ([]byte, error) {
	cmd := string(incomingCmd)

	if cmd == "*1\r\n$5\r\nmulti\r\n*2\r\n$3\r\nget\r\n$4\r\nname\r\n*1\r\n$4\r\nexec\r\n" {
		return []byte("+OK\r\n+QUEUED\r\n*1\r\n$3\r\nIBM\r\n"), nil
	}

	if cmd == "*1\r\n$5\r\nmulti\r\n*3\r\n$3\r\nset\r\n$4\r\nname\r\n$3\r\nIBM\r\n*1\r\n$4\r\nexec\r\n" {
		return []byte("+OK\r\n+QUEUED\r\n*1\r\n+OK\r\n"), nil
	}

	if cmd == "*1\r\n$5\r\nmulti\r\n*2\r\n$3\r\ndel\r\n$4\r\nname\r\n*1\r\n$4\r\nexec\r\n" {
		return []byte("+OK\r\n+QUEUED\r\n*1\r\n:1\r\n"), nil
	}

	if strings.Contains(cmd, "\r\nmulti\r\n") {
		return []byte("+OK\r\n+QUEUED\r\n+QUEUED\r\n+QUEUED\r\n*3\r\n+OK\r\n$3\r\nIBM\r\n:1\r\n"), nil
	}

	if isSingleRedis {
		reply := make([]byte, len(incomingCmd))
		copy(reply, incomingCmd)
		return reply, nil
	}

	if strings.Contains(cmd, "get-master-addr-by-name") {
		return []byte("*2\r\n$9\r\n127.0.0.1\r\n$4\r\n6382\r\n"), nil
	}

	if strings.Contains(cmd, "sentinels") {
		return []byte("*0\r\n"), nil
	}

	// Used by FailoverClusterClient
	if strings.Contains(cmd, "sentinel\r\n$6\r\nslaves") {
		return []byte("*0\r\n"), nil
	}

	// Case of Pipeline (without multi/exec commands)
	if cmd == "*3\r\n$3\r\nset\r\n$4\r\nname\r\n$3\r\nIBM\r\n*2\r\n$3\r\nget\r\n$4\r\nname\r\n*2\r\n$3\r\ndel\r\n$4\r\nname\r\n" {
		return []byte("+OK\r\n+QUEUED\r\n+QUEUED\r\n+QUEUED\r\n*3\r\n+OK\r\n$3\r\nIBM\r\n:1\r\n"), nil
	}

	if strings.Contains(cmd, "set") {
		return []byte("+OK\r\n"), nil
	}

	if strings.Contains(cmd, "get") {
		return []byte("$3\r\nIBM\r\n"), nil
	}

	if strings.Contains(cmd, "del") {
		return []byte(":1\r\n"), nil
	}

	// This commands are used by Cluster servers
	// For every Redis command you wanna test, you need to tell go-redis that the command is available.
	// The best way of listing them is by running a Redis server locally, connecting into them via telnet and running the "command" command.
	if strings.Contains(cmd, "command") {
		return []byte("*2\r\n*7\r\n$3\r\nget\r\n*7\r\n$3\r\nset\r\n:2\r\n*2\r\n+readonly\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*3\r\n+@read\r\n+@string\r\n+@fast\r\n"), nil
	}

	if strings.Contains(cmd, "cluster\r\n$5\r\nslots") {
		return []byte("*1\r\n*3\r\n:0\r\n:5460\r\n*3\r\n$9\r\n127.0.0.1\r\n:6382\r\n$40\r\nd31ffce4060efa89b6b6dac3da472ce03054e1a2\r\n"), nil
	}

	return nil, fmt.Errorf("Command not handled: '%s'", strings.Replace(cmd, "\r\n", "\\r\\n", -1))
}

func (c *MockConn) Read(b []byte) (n int, err error) {
	if c.clientType == Single {
		parsedCmd, err := parseIncomingCommand(c.connData, true)

		if err != nil {
			return 0, err
		}

		copy(b, parsedCmd)
		_len := len(parsedCmd)
		c.connData = []byte{}
		return _len, nil
	}

	// This case tests single redis with sentinel (NewFailoverClient), cluster (NewClusterClient) or ClusterFailover
	parsedCmd, err := parseIncomingCommand(c.connData, false)

	if err != nil {
		return 0, err
	}

	copy(b, parsedCmd)
	_len := len(parsedCmd)
	c.connData = []byte{}
	return _len, nil
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
			t.Run(fmt.Sprintf("%s - %s", redisTypeMap[rType], name), func(t *testing.T) {
				recorder := instana.NewTestRecorder()
				sensor := instana.NewSensorWithTracer(
					instana.NewTracerWithEverything(instana.DefaultOptions(), recorder),
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
						pipe.Exec(ctx)
					} else if len(example.DoTxPipeCommand) > 0 {
						pipe := rdb.TxPipeline()

						for _, doCmd := range example.DoTxPipeCommand {
							pipe.Do(ctx, doCmd...)
						}
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
						pipe.Exec(ctx)
					} else if len(example.DoTxPipeCommand) > 0 {
						pipe := rdb.TxPipeline()

						for _, doCmd := range example.DoTxPipeCommand {
							pipe.Do(ctx, doCmd...)
						}
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
