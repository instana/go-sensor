![golang banner 2017-03-15](https://cloud.githubusercontent.com/assets/395132/23948351/f360d228-0982-11e7-9b17-6a6146207ded.png)

# Instana Go Sensor
golang-sensor requires Go version 1.7 or greater.

The Instana Go sensor consists of two parts:

* metrics sensor
* [OpenTracing](http://opentracing.io) tracer

## Sensor

To use sensor only without tracing ability, import the `instana` package and run

	instana.InitSensor(opt)

in your main function. The init function takes an `Options` object with the following optional fields:

* **Service** - global service name that will be used to identify the program in the Instana backend
* **AgentHost**, **AgentPort** - default to localhost:42699, set the coordinates of the Instana proxy agent
* **LogLevel** - one of Error, Warn, Info or Debug

Once initialised, the sensor will try to connect to the given Instana agent and in case of connection success will send metrics and snapshot information through the agent to the backend.

## OpenTracing

In case you want to use the OpenTracing tracer, it will automatically initialise the sensor and thus also activate the metrics stream. To activate the global tracer, run for example

	ot.InitGlobalTracer(instana.NewTracerWithOptions(&instana.Options{
		Service:  SERVICE,
		LogLevel: instana.DEBUG}))

in your main functions. The tracer takes same options that the sensor takes for initialisation, described above.

The tracer is able to protocol and piggyback OpenTracing baggage, tags and logs. Only text mapping is implemented yet, binary is not supported. Also, the tracer tries to map the OpenTracing spans to the Instana model based on OpenTracing recommended tags. See `simple` example for details on how recommended tags are used.

The Instana tracer will remap OpenTracing HTTP headers into Instana Headers, so parallel use with some other OpenTracing model is not possible. The instana tracer is based on the OpenTracing Go basictracer with necessary modifications to map to the Instana tracing model. Also, sampling isn't implemented yet and will be focus of future work.

## Events API

The sensor, be it instantiated explicitly or implicitly throught the tracer, provides a simple wrapper API to send events into Instana as described in https://instana.atlassian.net/wiki/display/DOCS/Event+SDK+REST+Web+Service :

	SentDefaultServiceEvent: send the event with default service name
	SendServiceEvent: send the event for a service named explicitly
	SendHostEvent: send an event that will not be assigned to a service, like in case of general events

## Examples

Following examples are included in the `examples` folder:

* **ot-simple/simple.go** - demoes generally how to use the tracer
* **webserver/http.go** - demoes how http server and client should be instrumented
* **rpc/rpc.go** - demoes the fallback to RPC
* **event/event.go** - demoes the event API
