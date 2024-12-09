# Example Usage of instacosmos

### Prerequisites

* Create a cosmos database named "trace-data" in your Azure Cosmos DB account with the container id "spans" and partition key
 "/SpanID"
* export the endpoint and key as environment variable
```sh
export COSMOS_CONNECTION_URL="<endpoint>"
export COSMOS_KEY="<key>"
```

# setup cosmos example
```sh
cd ./examples/cosmos
go mod tidy
go run .
```

### Test the application
```sh
# Test success
curl http://localhost:9990/cosmos-test

# Test error
curl http://localhost:9990/cosmos-test?error=true
```
