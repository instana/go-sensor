An Instrumented Server gin Example
==========================

An example of usage of Instana tracer to instrumenta a Gin application with `https://github.com/instana/go-sensor/tree/master/instrumentation/instagin`.

Usage
-----

To start an example locally on `localhost:8881`.

```bash
go run main.go
```

In case when the port is already in use, please choose another one.

```
  -address string
        address to use by an example (default "localhost")
  -port string
        port to use by an example (default "8881")
```

There will be two endpoints exposed.

```
GET    /myendpoint               
GET    /v1/myendpoint    
```       
