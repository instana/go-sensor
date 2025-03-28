# Example usage of mongo-driver v2 instrumentation

## Set up

### Start MongoDB using Docker Compose:

```sh
docker-compose up -d
```

### Run your Go application:

```sh
go run main.go
```


## Test the application:

#### Insert an item:

```sh
curl -X POST -H "Content-Type: application/json" -d '{"name": "sample-name"}' http://localhost:8080/insert
```

NOTE: change value of name each time `/insert` API is called as an unique index is created on the field `name`.

#### Retrieve items:

```sh
curl http://localhost:8080/get
```

#### Simulate error

```sh
curl -X POST http://localhost:8080/error
```