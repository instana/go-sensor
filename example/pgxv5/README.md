
Student Database Service
----
This application demonstrates the instrumentation capabilities of `instapgxv2`, which is an instrumentation package for the 
[pgx/v5](https://github.com/jackc/pgx) library. Also, this example showcases how trace propagation occurs from an instrumented HTTP query to the database 
calls. Comments have been added in relevant places and that should provide sufficient guidance for using the instrumentation library. 


## Running the application
- An Instana Host agent must be running to collect the traces.
- Run `docker-compose up` from the `go-sensor` root folder. This will bring the database service up.
- Compile the example by `go build -o server .` and run the application using `./server`


## Querying the server
The available routes are,
- localhost:8080/insert
- localhost:8080/delete
- localhost:8080/

You can insert a record to the database by issuing the following curl request,
```
curl -X POST http://localhost:8080/insert \
-H "Content-Type: application/json" \
-d '{"studentname": "Liam Watson"}'
```
You can delete a record to the database by issuing the following curl request,
```
curl -X DELETE http://localhost:8080/delete/<id>
```

After issuing a couple of queries, you will be able to see the call traces in the Instana dashboard.

### Notes:
Additionally, one can connect to the database using the `psql` tool, by issuing the following command,
```
psql -h localhost -p 5434 -U pgxadmin -d students
```
and provide the password: `pgxpwd`
