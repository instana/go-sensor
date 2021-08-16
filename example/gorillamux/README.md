An Instrumented Server gorilla mux Example
==========================

An example of usage of Instana tracer to instrument a gorilla mux application with `https://github.com/instana/go-sensor/tree/master/instrumentation/instamux`.

Usage
-----

To start an example locally on `localhost:8881` run:

```bash
go run main.go -l localhost:8881
```

In case when the port is already in use, please choose another one.

```
  -l string
        Server listen address
```

There will be an exposed endpoint:

```
GET    /foo               
```       
