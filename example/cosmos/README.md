# Example Usage of instacosmos

### Prerequisites

* Create a cosmos database named "trace-data" in your Azure Cosmos DB account.
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
```
