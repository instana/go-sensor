// (c) Copyright IBM Corp. 2025

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/godror/godror"
	instana "github.com/instana/go-sensor"
	"github.com/jmoiron/sqlx"
)

var s instana.TracerLogger

func init() {
	s = instana.InitCollector(&instana.Options{
		Service:  "Oracle app",
		Tracer:   instana.DefaultTracerOptions(),
		LogLevel: instana.Info,
	})
}

func getTNSConnString() string {
	connStr := os.Getenv("ORACLE_CONNECTION_STRING")
	if connStr == "" {
		user := getEnvOrDefault("ORACLE_USER", "scott")
		password := getEnvOrDefault("ORACLE_PASSWORD", "tiger")

		tnsDescriptor := `(description=(CONNECT_TIMEOUT=40)(RETRY_COUNT=10)(TRANSPORT_CONNECT_TIMEOUT=3)` +
			`(address_list=(address=(protocol=tcp)(host=hostdb1)(port=1521))` +
			`(address=(protocol=tcp)(host=hostdb2)(port=1521))` +
			`(failover=on)(load_balance=off))` +
			`(connect_data=(service_name=ods-domain)))`

		connStr = fmt.Sprintf("%s/%s@%s", user, password, tnsDescriptor)
	}

	return connStr
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func initDbConn(connStr string) (*sqlx.DB, error) {
	driverName := "oracle"

	P, err := godror.ParseDSN(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DSN: %w", err)
	}

	gDrv := godror.NewConnector(P).Driver()
	instana.InstrumentSQLDriver(s.LegacySensor(), driverName, gDrv)

	dbx, err := sqlx.Open(driverName+"_with_instana", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection: %w", err)
	}

	dbx.SetMaxOpenConns(10)
	dbx.SetMaxIdleConns(5)
	dbx.SetConnMaxLifetime(time.Hour)
	dbx.SetConnMaxIdleTime(10 * time.Minute)

	ctx, cancel := context.WithTimeout(context.TODO(), 15*time.Second)
	defer cancel()

	if err = dbx.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	fmt.Println("Successfully connected to Oracle database")
	return dbx, nil
}

var dbConn *sqlx.DB

func handler(w http.ResponseWriter, req *http.Request) {
	var result string
	err := dbConn.QueryRowContext(req.Context(), `SELECT 'Connected to Oracle!' FROM DUAL`).Scan(&result)
	if err != nil {
		http.Error(w, fmt.Sprintf("Query error: %v", err), http.StatusInternalServerError)
		return
	}

	_, _ = fmt.Fprintf(w, "Result: %s", result)
}

func healthHandler(w http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), 2*time.Second)
	defer cancel()

	if err := dbConn.PingContext(ctx); err != nil {
		http.Error(w, "Database unhealthy", http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}

func main() {
	connStr := getTNSConnString()
	fmt.Println("Connecting to Oracle...")

	var err error
	dbConn, err = initDbConn(connStr)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize database connection: %v", err))
	}
	defer dbConn.Close()

	fmt.Println("Waiting for Instana agent...")
	for i := 0; i < 30; i++ {
		if instana.Ready() {
			fmt.Println("Instana agent ready")
			break
		}
		time.Sleep(1 * time.Second)
	}

	http.HandleFunc("/oracle", instana.TracingHandlerFunc(s, "/oracle", handler))
	http.HandleFunc("/health", healthHandler)

	fmt.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
