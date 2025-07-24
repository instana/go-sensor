// (c) Copyright IBM Corp. 2024

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instapgx/v2"
	"github.com/jackc/pgx/v5"
)

const DB_NAME string = "postgres"
const DB_USER string = "postgres"
const DB_PASS string = "mysecretpassword"
const DB_HOST string = "localhost"
const DB_PORT string = "5432"

var TableName = "deutsch"

func main() {
	TableName = TableName + "-" + strconv.Itoa(rand.Intn(100000))

	// Initialising Instana instrumentation collector
	col := instana.InitCollector(&instana.Options{
		Service: "testwebapp",
		Tracer:  instana.DefaultTracerOptions(),
	})

	// Setting up an instrumented database connection
	var conn *pgx.Conn
	var err error
	// We'll use the Instana collector to set up an instrumented database connection
	if conn, err = initDatabase(DB_NAME, TableName, col); err != nil {
		log.Fatal("init db:", err.Error())
	}

	// Setting up instrumented handler for different routes
	r := mux.NewRouter()
	setupRoutes(r, conn, col)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// Channel to listen for termination signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	log.Println("Starting server...")
	go func() {
		err = srv.ListenAndServe()
		if err != nil {
			log.Printf("http listen: %s", err.Error())
		}
	}()

	// Block until a signal is received
	<-quit
	log.Println("Shutting down server...")

	// Create a context with a timeout to allow for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Shutdown the server gracefully
	if err := exitGracefully(ctx, conn, srv); err != nil {
		log.Println("Server forced to shutdown: %s", err)
	}

	log.Println("Server exiting")
}

func initDatabase(db string, table string, tr instana.TracerLogger) (*pgx.Conn, error) {
	user := DB_USER
	pwd := DB_PASS
	host := DB_HOST
	port := DB_PORT

	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pwd, host, port, db)
	cfg, err := pgx.ParseConfig(dbUrl)
	if err != nil {
		return nil, err
	}

	// Here we are assigning the Instana tracer for pgx/v5 to the Tracer interface
	cfg.Tracer = instapgx.InstanaTracer(cfg, tr)

	conn, err := pgx.ConnectConfig(context.Background(), cfg)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (StudentID SERIAL PRIMARY KEY, studentname VARCHAR(50))`, pgx.Identifier{table}.Sanitize())
	_, err = conn.Exec(context.Background(), query)
	if err != nil {
		log.Fatal("failed to create table:", err.Error())
	}

	return conn, nil
}

func setupRoutes(r *mux.Router, conn *pgx.Conn, tr instana.TracerLogger) {
	// We use Instana HTTP instrumentation APIs for instrumenting the handlers
	r.HandleFunc("/", instana.TracingHandlerFunc(tr, "/", handleHome))
	r.HandleFunc("/insert", instana.TracingHandlerFunc(tr, "/insert", handleInsert(conn))).Methods("POST")
	r.HandleFunc("/delete/{id}", instana.TracingHandlerFunc(tr, "/delete/{id}", handleDelete(conn))).Methods("DELETE")
}

func exitGracefully(ctx context.Context, conn *pgx.Conn, srv *http.Server) error {
	query := fmt.Sprintf(`DROP TABLE IF EXISTS %s`, pgx.Identifier{TableName}.Sanitize())
	_, err := conn.Exec(context.Background(), query)
	if err != nil {
		log.Printf("\nExec failed: %v\n", err.Error())
		return err
	}
	log.Printf("\ndrop on table %s succeeded\n", TableName)

	err = conn.Close(context.Background())
	if err != nil {
		log.Println("close db:", err.Error())
	}

	if err = srv.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}

func handleHome(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("Hello, World!"))
	if err != nil {
		log.Println("write response:", err.Error())
	}
}

func handleDelete(conn *pgx.Conn) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		query := fmt.Sprintf("DELETE FROM %s WHERE StudentID = $1", pgx.Identifier{TableName}.Sanitize())

		// IMPORTANT: Remember to pass the context of the HTTP request to ensure trace propagation to the database calls.
		ct, err := conn.Exec(r.Context(), query, id)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write([]byte(err.Error()))
			if err != nil {
				log.Println("write response:", err.Error())
			}
			return
		}
		if ct.RowsAffected() == 0 {
			w.WriteHeader(http.StatusNotFound)
			_, err := w.Write([]byte("student not found"))
			if err != nil {
				log.Println("write response:", err.Error())
			}

		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(fmt.Sprintf("Deleted student id %s", id)))
		if err != nil {
			log.Println("write response:", err.Error())
		}
	}
}

func handleInsert(conn *pgx.Conn) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		var reqBody struct {
			StudentName string `json:"studentname"`
		}

		err := json.NewDecoder(r.Body).Decode(&reqBody)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, err = w.Write([]byte(err.Error()))
			if err != nil {
				log.Println("write response:", err.Error())
			}
			return
		}

		query := fmt.Sprintf(`INSERT INTO %s(studentname) VALUES ($1)`, pgx.Identifier{TableName}.Sanitize())

		// IMPORTANT: Remember to pass the context of the HTTP request to ensure trace propagation to the database calls.
		_, err = conn.Exec(r.Context(), query, reqBody.StudentName)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, err = w.Write([]byte(err.Error()))
			if err != nil {
				log.Println("write response:", err.Error())
			}
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte("A record was inserted."))
		if err != nil {
			log.Println("write response:", err.Error())
		}
	}
}
