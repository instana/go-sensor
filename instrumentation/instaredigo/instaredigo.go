//go:build go1.16
// +build go1.16

package instaredigo

import (
	"context"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"
)

type instaRedigoConn struct {
	redis.Conn
	sensor   *instana.Sensor
	address  string
	prevSpan *prevSpan
}

type prevSpan struct {
	span      ot.Span
	batchCmds []string
}

type instaRedigoPool struct {
    redis.Pool
    sensor *instana.Sensor
}

//GetContext gets a connection using the provided context with the instana sensor.
func (p *instaRedigoPool) GetContext(s *instana.Sensor, ctx  context.Context) (redis.Conn, error) {
    conn,err := p.Pool.GetContext(ctx)
    if err != nil {
        return conn, err
    } else {
        return &instaRedigoConn{Conn: conn, sensor: s, address: "", prevSpan: nil}, err
    }
}

//Get gets a connection with the instana sensor.
func (p *instaRedigoPool) Get(s *instana.Sensor) redis.Conn {
    conn := p.Pool.Get()
    return &instaRedigoConn{Conn: conn, sensor: s, address: "", prevSpan: nil}
}

//DoContext sends a command to server and returns the received reply along with 
//the instrumentation.
func (c *instaRedigoConn) DoContext(ctx context.Context, commandName string,
	args ...interface{}) (reply interface{}, err error) {
	if cwc, ok := c.Conn.(redis.ConnWithContext); ok {
		if c.prevSpan != nil {
			if strings.ToUpper(commandName) == "EXEC" || strings.ToUpper(commandName) == "DISCARD" {
				reply, err = cwc.DoContext(ctx, commandName, args...)
				if err != nil {
					c.prevSpan.span.SetTag("redis.error", err.Error())
					c.prevSpan.span.LogFields(otlog.Object("error", err.Error()))
				}
				c.prevSpan.span.Finish()
				c.prevSpan = nil
			} else {
				c.prevSpan.batchCmds = append(c.prevSpan.batchCmds, commandName)
				c.prevSpan.span.SetTag("redis.subCommands", c.prevSpan.batchCmds)
				reply, err = cwc.DoContext(ctx, commandName, args...)
				if err != nil {
					c.prevSpan.span.SetTag("redis.error", err.Error())
					c.prevSpan.span.LogFields(otlog.Object("error", err.Error()))
				}
			}
		} else {
			tracer := c.sensor.Tracer()
			opts := []ot.StartSpanOption{
				ot.Tags{
					"redis.connection": c.address,
				},
			}
			if ps, ok := instana.SpanFromContext(ctx); ok {
				tracer = ps.Tracer()
				opts = append(opts, ot.ChildOf(ps.Context()))
			}
			span := tracer.StartSpan("redis", opts...)
			if strings.ToUpper(commandName) == "MULTI" {
				c.prevSpan = &prevSpan{span, nil}
			}
			span.SetTag("redis.command", commandName)
			reply, err = cwc.DoContext(ctx, commandName, args...)
			if err != nil {
				span.SetTag("redis.error", err.Error())
				span.LogFields(otlog.Object("error", err.Error()))
			}
			if strings.ToUpper(commandName) != "MULTI" {
				span.Finish()
			}
		}
		return reply, err
	}
	return nil, errors.New("redis: connection does not support ConnWithContext")
}

//DoWithTimeout executes a Redis command with the specified read timeout along 
//with the instrumentation. If the connection does not satisfy the ConnWithTimeout 
//interface, then an error is returned.
func (c *instaRedigoConn) DoWithTimeout(timeout time.Duration, commandName string, 
    args ...interface{}) (reply interface{}, err error) {
	if cwt, ok := c.Conn.(redis.ConnWithTimeout); ok {
		//return cwt.DoWithTimeout(timeout, commandName, args...)
		var cmdArgs []interface{}
		ctx := context.Background()
		//Separating the passed context and the arguments
		for _, a := range args {
			if _, ok := a.(context.Context); ok {
				ctx = a.(context.Context)
			} else {
				cmdArgs = append(cmdArgs, a)
			}
		}
		if c.prevSpan != nil {
			if strings.ToUpper(commandName) == "EXEC" || strings.ToUpper(commandName) == "DISCARD" {
				reply, err = cwt.DoWithTimeout(timeout, commandName, cmdArgs...) 
				if err != nil {
					c.prevSpan.span.SetTag("redis.error", err.Error())
					c.prevSpan.span.LogFields(otlog.Object("error", err.Error()))
				}
				c.prevSpan.span.Finish()
				c.prevSpan = nil
			} else {
				c.prevSpan.batchCmds = append(c.prevSpan.batchCmds, commandName)
				c.prevSpan.span.SetTag("redis.subCommands", c.prevSpan.batchCmds)
				reply, err = cwt.DoWithTimeout(timeout, commandName, cmdArgs...) 
				if err != nil {
					c.prevSpan.span.SetTag("redis.error", err.Error())
					c.prevSpan.span.LogFields(otlog.Object("error", err.Error()))
				}
			}
		} else {
			tracer := c.sensor.Tracer()
			opts := []ot.StartSpanOption{
				ot.Tags{
					"redis.connection": c.address,
				},
			}
			if ps, ok := instana.SpanFromContext(ctx); ok {
				tracer = ps.Tracer()
				opts = append(opts, ot.ChildOf(ps.Context()))
			}
			span := tracer.StartSpan("redis", opts...)
			if strings.ToUpper(commandName) == "MULTI" {
				c.prevSpan = &prevSpan{span, nil}
			}
			span.SetTag("redis.command", commandName)
			reply, err = cwt.DoWithTimeout(timeout, commandName, cmdArgs...) 
			if err != nil {
				span.SetTag("redis.error", err.Error())
				span.LogFields(otlog.Object("error", err.Error()))
			}
			if strings.ToUpper(commandName) != "MULTI" {
				span.Finish()
			}
		}
		return reply, err
	}
	return nil, errors.New("redis: connection does not support ConnWithTimeout")
}

//Dial connects to the Redis server at the given network and address using the
//specified options along with instrumentation code.
func Dial(sensor *instana.Sensor, network, address string, options ...redis.DialOption) (redis.Conn, error) {
	conn, err := redis.Dial(network, address, options...)
	if strings.HasPrefix(address, ":") {
		address = "localhost" + address
	}
	if err != nil {
		return conn, err
	} else {
		return &instaRedigoConn{Conn: conn, sensor: sensor, address: address, prevSpan: nil}, err
	}
}

//DialURL wraps DialURLContext using context.Background along with the instrumentation code.
func DialURL(sensor *instana.Sensor, rawurl string, options ...redis.DialOption) (redis.Conn, error) {
	conn, err := redis.DialURL(rawurl, options...)
	if err != nil {
		return conn, err
	} else {
		return &instaRedigoConn{conn, sensor, rawurl, nil}, err
	}
}

//DialURLContext connects to a Redis server at the given URL using the Redis
//URI scheme along with the instrumentation code.
func DialURLContext(sensor *instana.Sensor, ctx context.Context, rawurl string, options ...redis.DialOption) (redis.Conn, error) {
	conn, err := redis.DialURLContext(ctx, rawurl, options...)
	if err != nil {
		return conn, err
	} else {
		return &instaRedigoConn{conn, sensor, rawurl, nil}, err
	}
}

//NewConn returns a new Redigo connection for the given net connection along with the instrumentation code.
func NewConn(sensor *instana.Sensor, netConn net.Conn, readTimeout, writeTimeout time.Duration) redis.Conn {
	addr := netConn.LocalAddr().String()
	conn := redis.NewConn(netConn, readTimeout, writeTimeout)
	return &instaRedigoConn{conn, sensor, addr, nil}
}

// Do sends a command to the server and returns the received reply and collect
//the instrumentation details. Inorder to capture the correlated span information,
//create a context from the parent span and pass it as an argument along with
//the other arguments. The Do API will retrieve the span information from the
//context and record the correlated span information.
func (c *instaRedigoConn) Do(commandName string, args ...interface{}) (reply interface{}, err error) {
	var cmdArgs []interface{}
	ctx := context.Background()
	//Separating the passed context and the arguments
	for _, a := range args {
		if _, ok := a.(context.Context); ok {
			ctx = a.(context.Context)
		} else {
			cmdArgs = append(cmdArgs, a)
		}
	}
	if c.prevSpan != nil {
		if strings.ToUpper(commandName) == "EXEC" || strings.ToUpper(commandName) == "DISCARD" {
			reply, err = c.Conn.Do(commandName, cmdArgs...)
			if err != nil {
				c.prevSpan.span.SetTag("redis.error", err.Error())
				c.prevSpan.span.LogFields(otlog.Object("error", err.Error()))
			}
			c.prevSpan.span.Finish()
			c.prevSpan = nil
		} else {
			c.prevSpan.batchCmds = append(c.prevSpan.batchCmds, commandName)
			c.prevSpan.span.SetTag("redis.subCommands", c.prevSpan.batchCmds)
			reply, err = c.Conn.Do(commandName, cmdArgs...)
			if err != nil {
				c.prevSpan.span.SetTag("redis.error", err.Error())
				c.prevSpan.span.LogFields(otlog.Object("error", err.Error()))
			}
		}
	} else {
		tracer := c.sensor.Tracer()
		opts := []ot.StartSpanOption{
			ot.Tags{
				"redis.connection": c.address,
			},
		}
		if ps, ok := instana.SpanFromContext(ctx); ok {
			tracer = ps.Tracer()
			opts = append(opts, ot.ChildOf(ps.Context()))
		}
		span := tracer.StartSpan("redis", opts...)
		if strings.ToUpper(commandName) == "MULTI" {
			c.prevSpan = &prevSpan{span, nil}
		}
		span.SetTag("redis.command", commandName)
		reply, err = c.Conn.Do(commandName, cmdArgs...)
		if err != nil {
			span.SetTag("redis.error", err.Error())
			span.LogFields(otlog.Object("error", err.Error()))
		}
		if strings.ToUpper(commandName) != "MULTI" {
			span.Finish()
		}
	}
	return reply, err
}

// Send writes the command to the client's output buffer and collect the
//instrumentation details.Inorder to capture the correlated span information,
//create a context from the parent span and pass it as an argument along with
//the other arguments. The Send API will retrieve the span information from the
//context and record the correlated span information.
func (c *instaRedigoConn) Send(commandName string, args ...interface{}) (err error) {
	var cmdArgs []interface{}
	//Separating the parent context from the arguments
	ctx := context.Background()
	for _, a := range args {
		if _, ok := a.(context.Context); ok {
			ctx = a.(context.Context)
		} else {
			cmdArgs = append(cmdArgs, a)
		}
	}
	if c.prevSpan != nil {
		if strings.ToUpper(commandName) == "EXEC" || strings.ToUpper(commandName) == "DISCARD" {
			err = c.Conn.Send(commandName, cmdArgs...)
			if err != nil {
				c.prevSpan.span.SetTag("redis.error", err.Error())
				c.prevSpan.span.LogFields(otlog.Object("error", err.Error()))
			}
			c.prevSpan.span.Finish()
			c.prevSpan = nil
		} else {
			c.prevSpan.batchCmds = append(c.prevSpan.batchCmds, commandName)
			c.prevSpan.span.SetTag("redis.subCommands", c.prevSpan.batchCmds)
			err = c.Conn.Send(commandName, cmdArgs...)
			if err != nil {
				c.prevSpan.span.SetTag("redis.error", err.Error())
				c.prevSpan.span.LogFields(otlog.Object("error", err.Error()))
			}
		}
	} else {
		tracer := c.sensor.Tracer()
		opts := []ot.StartSpanOption{
			ot.Tags{
				"redis.connection": c.address,
			},
		}
		if ps, ok := instana.SpanFromContext(ctx); ok {
			tracer = ps.Tracer()
			opts = append(opts, ot.ChildOf(ps.Context()))
		}
		span := tracer.StartSpan("redis", opts...)
		if strings.ToUpper(commandName) == "MULTI" {
			c.prevSpan = &prevSpan{span, nil}
		}
		span.SetTag("redis.command", commandName)
		err = c.Conn.Send(commandName, cmdArgs...)
		if err != nil {
			span.SetTag("redis.error", err.Error())
			span.LogFields(otlog.Object("error", err.Error()))
		}
		if strings.ToUpper(commandName) != "MULTI" {
			span.Finish()
		}
	}
	return err
}
