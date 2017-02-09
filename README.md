# Instana Go Sensor

The Instana Go sensor consists of two parts:

* metrics sensor
* [OpenTracing](http://opentracing.io) tracer

To use sensor only without tracing ability, import the `instana` package and run

	instana.InitSensor(opt)

in your main function. The init function takes an `Options` object with the following optional fields:

* **Service** - global service name that will be used to identify the program in the Instana backend
* **AgentHost**, **AgentPort** - default to localhost:42699, set the coordinates of the Instana proxy agent
* **LogLevel** - one of Error, Warn, Info or Debug

Once initialised, the sensor will try to connect to the given Instana agent and in case of connection success will send metrics and snapshot information through the agent to the backend.

In case you want to use the OpenTracing tracer, it will automatically initialise the sensor and thus also activate the metrics stream. To activate the global tracer, run for example


ot.InitGlobalTracer(instana.NewTracerWithOptions(
    &instana.Options{ Service:  SERVICE }
))

in your main functions. The tracer takes same options that the sensor takes for initialisation, described above.

The tracer is able to protocol and piggyback OpenTracing baggage, tags and logs. Only text mapping is implemented yet, binary is not supported. Also, the tracer tries to map the OpenTracing spans to the Instana model based on OpenTracing recommended tags. See `simple` example for details on how recommended tags are used.

The Instana tracer will remap OpenTracing HTTP headers into Instana Headers, so parallel use with some other OpenTracing model is not possible. The instana tracer is based on the OpenTracing Go basictracer with necessary modifications to map to the Instana tracing model. Also, sampling isn't implemented yet and will be focus of future work.

Following examples are included in the `examples` folder:

* **simple** - demoes generally how to use the tracer
* **http** - demoes how http server and client should be instrumented
* **rpc** - demoes the fallback to RPC

In case you'd like to debug or you need to troubleshoot, you can set `instana.Logger` and capture the output.
