APP_PORT=:8134

start-docker:
	docker run --name http-database-greeter-postgres -e POSTGRES_PASSWORD=mysecretpassword -p 5432:5432 -d postgres

remove-docker:
	docker rm -f http-database-greeter-postgres

start-app:
	go run . -db postgres://postgres:mysecretpassword@localhost:5432/postgres?sslmode=disable -l $(APP_PORT)

call:
	curl -D- http://localhost$(APP_PORT)

clean: remove-docker