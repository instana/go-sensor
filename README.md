![golang banner 2017-07-11](https://disznc.s3.amazonaws.com/Instana-Go-2017-07-11-at-16.01.45.png)

# Instana Go Sensor
go-sensor requires Go version 1.7 or greater.

The Instana Go sensor consists of two parts:

* metrics sensor
* [OpenTracing](http://opentracing.io) tracer

[![Build Status](https://travis-ci.org/instana/go-sensor.svg?branch=master)](https://travis-ci.org/instana/go-sensor)
[![OpenTracing Badge](https://img.shields.io/badge/OpenTracing-enabled-blue.svg)](http://opentracing.io)

## Common Operations

The Instana Go sensor offers a set of quick features to support tracing of the most common operations like handling HTTP requests and executing HTTP requests.

To create an instance of the Instana sensor just request a new instance using the _instana.NewSensor_ factory method and providing the name of the application. It is recommended to use a single Instana only. The sensor implementation is fully thread-safe and can be shared by multiple threads.

```
var sensor = instana.NewSensor("my-service")
```

A full example can be found under the examples folder in _example/webserver/instana/http.go_.

### HTTP Server Handlers

With support to wrap a _http.HandlerFunc_, Instana quickly adds the possibility to trace requests and collect child spans, executed in the context of the request span.

Minimal changes are required for Instana to be able to capture the necessary information. By simply wrapping the currently existing _http.HandlerFunc_ Instana collects and injects necessary information automatically.

That said, a simple handler function like the following will simple be wrapped and registered like normal.

For your own preference registering the handler and wrapping it can be two separate steps or a single one. The following example code shows both versions, starting with two steps.
```
func myHandler(w http.ResponseWriter, req *http.Request) {
  time.Sleep(450 * time.Millisecond)
}

// Doing registration and wrapping in two separate steps
func main() {
  http.HandleFunc(
      "/path/to/handler",
      sensor.TracingHandler("myHandler", myHandler),
  )
}

// Doing registration and wrapping in a single step
func main() {
  http.HandleFunc(
      sensor.TraceHandler("myHandler", "/path/to/handler", myHandler),
  )
}
```  

### Executing HTTP Requests

Requesting data or information from other, often external systems, is commonly implemented through HTTP requests. To make sure traces contain all spans, especially over all the different systems, certain span information have to be injected into the HTTP request headers before sending it out. Instana's Go sensor provides support to automate this process as much as possible.

To have Instana inject information into the request headers, create the _http.Request_ as normal and wrap it with the Instana sensor function as in the following example. 

```
req, err := http.NewRequest("GET", url, nil)
client := &http.Client{}
resp, err := sensor.TracingHttpRequest(
    "myExternalCall",
    parentRequest,
    req,
    client
)
```

The provided _parentRequest_ is the incoming request from the request handler (see above) and provides the necessary tracing and span information to create a child span and inject it into the request.

The request is, after injection, executing using the provided _http.Client_ instance. Like the normal _client.Do_ operation, the call will return a _http.Response_ instance or an error proving information of the failure reason.

## Sensor

To use sensor only without tracing ability, import the `instana` package and run

```Go
instana.InitSensor(opt)
```

in your main function. The init function takes an `Options` object with the following optional fields:

* **Service** - global service name that will be used to identify the program in the Instana backend
* **AgentHost**, **AgentPort** - default to localhost:42699, set the coordinates of the Instana proxy agent
* **LogLevel** - one of Error, Warn, Info or Debug

Once initialized, the sensor will try to connect to the given Instana agent and in case of connection success will send metrics and snapshot information through the agent to the backend.

## OpenTracing

In case you want to use the OpenTracing tracer, it will automatically initialize the sensor and thus also activate the metrics stream. To activate the global tracer, run for example

```Go
ot.InitGlobalTracer(instana.NewTracerWithOptions(&instana.Options{
	Service:  SERVICE,
	LogLevel: instana.DEBUG}))
```

in your main function. The tracer takes the same options that the sensor takes for initialization, described above.

The tracer is able to protocol and piggyback OpenTracing baggage, tags and logs. Only text mapping is implemented yet, binary is not supported. Also, the tracer tries to map the OpenTracing spans to the Instana model based on OpenTracing recommended tags. See `simple` example for details on how recommended tags are used.

The Instana tracer will remap OpenTracing HTTP headers into Instana Headers, so parallel use with some other OpenTracing model is not possible. The Instana tracer is based on the OpenTracing Go basictracer with necessary modifications to map to the Instana tracing model. Also, sampling isn't implemented yet and will be focus of future work.

## Events API

The sensor, be it instantiated explicitly or implicitly through the tracer, provides a simple wrapper API to send events to Instana as described in [its documentation](https://docs.instana.io/quick_start/api/#event-sdk-rest-web-service).

To learn more, see the [Events API](https://github.com/instana/go-sensor/blob/master/EventAPI.md) document in this repository.

## Examples

Following examples are included in the `examples` folder:

* **ot-simple/simple.go** - Demonstrates basic usage of the tracer
* **webserver/http.go** - Demonstrates how http server and client should be instrumented
* **rpc/rpc.go** - Demonstrates a basic RPC service
* **event/** - Demonstrates usage of the event API
