// (c) Copyright IBM Corp. 2024

//go:build integration
// +build integration

package instapgx

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"testing"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
)

const DB_URL = "postgres://postgres:mysecretpassword@localhost:5432/postgres?sslmode=disable"

func TestQueryAPI(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	sensor := instana.NewSensorWithTracer(tracer)

	parentSpan := sensor.StartSpan("test-service")
	defer parentSpan.Finish()

	cfg, err := pgx.ParseConfig(DB_URL)
	if err != nil {
		t.Error("unable to create pgx cfg")
	}

	cfg.Tracer = InstanaTracer(cfg, sensor)
	ctx := instana.ContextWithSpan(context.Background(), parentSpan)
	conn, err := pgx.ConnectConfig(ctx, cfg)
	if err != nil {
		t.Errorf("unable to connect to database: %v\n", err)
	}
	defer conn.Close(ctx)

	tableName := "student-" + strconv.Itoa(rand.Intn(90000)+10000)
	df := createTable(ctx, conn, tableName)
	defer df()

	var name string
	query := fmt.Sprintf(`SELECT studentname FROM %s`, pgx.Identifier{tableName}.Sanitize())
	if err := conn.QueryRow(ctx, query).Scan(&name); err != nil {
		fmt.Fprintf(os.Stderr, "queryRow failed: %v\n", err)
	}

	spans := recorder.GetQueuedSpans()
	assert.Len(t, spans, 4)
}

func TestExecAPI(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	sensor := instana.NewSensorWithTracer(tracer)

	parentSpan := sensor.StartSpan("test-service")
	defer parentSpan.Finish()

	cfg, err := pgx.ParseConfig(DB_URL)
	if err != nil {
		t.Error("unable to create pgx cfg")
	}

	cfg.Tracer = InstanaTracer(cfg, sensor)

	ctx := instana.ContextWithSpan(context.Background(), parentSpan)
	conn, err := pgx.ConnectConfig(ctx, cfg)

	if err != nil {
		t.Errorf("unable to connect to database: %v\n", err)
	}
	defer conn.Close(ctx)

	tableName := "student-" + strconv.Itoa(rand.Intn(90000)+10000)
	df := createTable(ctx, conn, tableName)
	defer df()

	sqlStmt := fmt.Sprintf(`INSERT INTO %s(studentname) VALUES ('Bala')`, pgx.Identifier{tableName}.Sanitize())
	_, err = conn.Exec(ctx, sqlStmt)
	if err != nil {
		fmt.Println("error observed for exec call")
	}

	spans := recorder.GetQueuedSpans()
	assert.Len(t, spans, 3)
}

func TestQueryRowAPI(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	sensor := instana.NewSensorWithTracer(tracer)

	parentSpan := sensor.StartSpan("test-service")
	defer parentSpan.Finish()

	cfg, err := pgx.ParseConfig(DB_URL)
	if err != nil {
		t.Error("unable to create pgx cfg")
	}

	cfg.Tracer = InstanaTracer(cfg, sensor)

	ctx := instana.ContextWithSpan(context.Background(), parentSpan)
	conn, err := pgx.ConnectConfig(ctx, cfg)

	if err != nil {
		t.Errorf("unable to connect to database: %v\n", err)
	}
	defer conn.Close(ctx)

	tableName := "student-" + strconv.Itoa(rand.Intn(90000)+10000)
	df := createTable(ctx, conn, tableName)
	defer df()

	var name string
	query := fmt.Sprintf(`SELECT studentname FROM %s`, pgx.Identifier{tableName}.Sanitize())
	err = conn.QueryRow(ctx, query).Scan(&name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "queryRow failed: %v\n", err)
	}

	spans := recorder.GetQueuedSpans()
	assert.Len(t, spans, 4)
}

func TestBeginAPI(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	sensor := instana.NewSensorWithTracer(tracer)

	parentSpan := sensor.StartSpan("test-service")
	defer parentSpan.Finish()

	cfg, err := pgx.ParseConfig(DB_URL)
	if err != nil {
		t.Error("unable to create pgx cfg")
	}

	cfg.Tracer = InstanaTracer(cfg, sensor)

	ctx := instana.ContextWithSpan(context.Background(), parentSpan)
	conn, err := pgx.ConnectConfig(ctx, cfg)

	if err != nil {
		t.Errorf("unable to connect to database: %v\n", err)
	}
	defer conn.Close(ctx)

	tableName := "student-" + strconv.Itoa(rand.Intn(90000)+10000)
	df := createTable(ctx, conn, tableName)
	defer df()

	tx, err := conn.Begin(ctx)
	if err != nil {
		log.Fatal("error while executing Begin")
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				log.Fatal("failed to rollback transaction", rollbackErr)
			}
		} else {
			if commitErr := tx.Commit(ctx); commitErr != nil {
				log.Fatal("failed to commit transaction", commitErr)
			}
		}
	}()

	insertQuery := fmt.Sprintf(`INSERT INTO %s(studentname) VALUES ('Malice')`, pgx.Identifier{tableName}.Sanitize())
	_, err = tx.Exec(ctx, insertQuery)
	if err != nil {
		fmt.Println("error occurred while INSERT statement")
	}

	updateQuery := fmt.Sprintf(`UPDATE %s SET studentname = 'Alice' WHERE studentname = 'Malice'`, pgx.Identifier{tableName}.Sanitize())
	_, err = tx.Exec(ctx, updateQuery)
	if err != nil {
		fmt.Println("error occurred while UPDATE statement")
	}

	spans := recorder.GetQueuedSpans()
	assert.Len(t, spans, 5)
}

func TestBeginTransactionAPI(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	sensor := instana.NewSensorWithTracer(tracer)

	parentSpan := sensor.StartSpan("test-service")
	defer parentSpan.Finish()

	cfg, err := pgx.ParseConfig(DB_URL)
	if err != nil {
		t.Error("unable to create pgx cfg")
	}

	cfg.Tracer = InstanaTracer(cfg, sensor)

	ctx := instana.ContextWithSpan(context.Background(), parentSpan)
	conn, err := pgx.ConnectConfig(ctx, cfg)

	if err != nil {
		t.Errorf("unable to connect to database: %v\n", err)
	}
	defer conn.Close(ctx)

	tableName := "student-" + strconv.Itoa(rand.Intn(90000)+10000)
	df := createTable(ctx, conn, tableName)
	defer df()

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		log.Fatal("error while executing Begin")
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				log.Fatal("failed to rollback transaction", rollbackErr)
			}
		} else {
			if commitErr := tx.Commit(ctx); commitErr != nil {
				log.Fatal("failed to commit transaction", commitErr)
			}
		}
	}()

	insertQuery := fmt.Sprintf(`INSERT INTO %s(studentname) VALUES ('Malice')`, pgx.Identifier{tableName}.Sanitize())
	_, err = tx.Exec(ctx, insertQuery)
	if err != nil {
		fmt.Println("error occurred while INSERT statement")
	}

	updateQuery := fmt.Sprintf(`UPDATE %s SET studentname = 'Alice' WHERE studentname = 'Malice'`, pgx.Identifier{tableName}.Sanitize())
	_, err = tx.Exec(ctx, updateQuery)
	if err != nil {
		fmt.Println("error occurred while UPDATE statement")
		return
	}

	spans := recorder.GetQueuedSpans()
	assert.Len(t, spans, 5)
}

func TestCopyFromAPI(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	sensor := instana.NewSensorWithTracer(tracer)

	parentSpan := sensor.StartSpan("test-service")
	defer parentSpan.Finish()

	cfg, err := pgx.ParseConfig(DB_URL)
	if err != nil {
		t.Error("unable to create pgx cfg")
	}

	cfg.Tracer = InstanaTracer(cfg, sensor)

	ctx := instana.ContextWithSpan(context.Background(), parentSpan)
	conn, err := pgx.ConnectConfig(ctx, cfg)

	if err != nil {
		t.Errorf("unable to connect to database: %v\n", err)
	}
	defer conn.Close(ctx)

	tableName := "student-" + strconv.Itoa(rand.Intn(90000)+10000)
	df := createTable(ctx, conn, tableName)
	defer df()

	rows := [][]interface{}{
		{"Kaushik"},
		{"Robert"},
		{"Tatya"},
	}

	columns := []string{"studentname"}

	_, err = conn.CopyFrom(ctx, pgx.Identifier{tableName}, columns, pgx.CopyFromRows(rows))
	if err != nil {
		log.Fatal("error while copying the data", err.Error())
	}

	spans := recorder.GetQueuedSpans()
	assert.Len(t, spans, 4)
}

func TestBatchAPI(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	sensor := instana.NewSensorWithTracer(tracer)

	parentSpan := sensor.StartSpan("test-service")
	defer parentSpan.Finish()

	cfg, err := pgx.ParseConfig(DB_URL)
	if err != nil {
		t.Error("unable to create pgx cfg")
	}

	cfg.Tracer = InstanaTracer(cfg, sensor)

	ctx := instana.ContextWithSpan(context.Background(), parentSpan)
	conn, err := pgx.ConnectConfig(ctx, cfg)

	if err != nil {
		t.Errorf("unable to connect to database: %v\n", err)
	}
	defer conn.Close(ctx)

	tableName := "student-" + strconv.Itoa(rand.Intn(90000)+10000)
	df := createTable(ctx, conn, tableName)
	defer df()

	batch := &pgx.Batch{}

	batch.Queue(fmt.Sprintf(`INSERT INTO %s(studentname) VALUES ('Alison')`, pgx.Identifier{tableName}.Sanitize()))
	batch.Queue(fmt.Sprintf(`INSERT INTO %s(studentname) VALUES ('Bobin')`, pgx.Identifier{tableName}.Sanitize()))

	br := conn.SendBatch(ctx, batch)

	_, err = br.Exec()
	if err != nil {
		t.Errorf("\nFailed to execute the batch commands")
	}

	err = br.Close()
	if err != nil {
		t.Errorf("\nFailed to close the batch results")
	}

	spans := recorder.GetQueuedSpans()
	assert.Len(t, spans, 7)
}

func TestInvalidConfig(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	sensor := instana.NewSensorWithTracer(tracer)

	var cfg *pgx.ConnConfig

	instanaTracer := InstanaTracer(cfg, sensor)
	assert.Nil(t, instanaTracer)
}

func createTable(ctx context.Context, conn *pgx.Conn, tableName string) func() {
	if conn == nil {
		return nil
	}
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (StudentID SERIAL PRIMARY KEY, studentname VARCHAR(50))`, pgx.Identifier{tableName}.Sanitize())
	_, err := conn.Exec(ctx, query)
	if err != nil {
		panic(err)
	}

	insertQuery := fmt.Sprintf(`INSERT INTO %s(studentname) VALUES ('Caleb'),('Liam');`, pgx.Identifier{tableName}.Sanitize())
	_, err = conn.Exec(ctx, insertQuery)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Exec failed: %v\n", err.Error())
	}

	dropTable := func() {
		query := fmt.Sprintf(`DROP TABLE IF EXISTS %s`, pgx.Identifier{tableName}.Sanitize())
		_, err := conn.Exec(ctx, query)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Exec failed: %v\n", err.Error())
		}
	}

	return dropTable
}

type alwaysReadyClient struct{}

func (alwaysReadyClient) Ready() bool                                { return true }
func (alwaysReadyClient) SendMetrics(_ acceptor.Metrics) error       { return nil }
func (alwaysReadyClient) SendEvent(_ *instana.EventData) error       { return nil }
func (alwaysReadyClient) SendSpans(_ []instana.Span) error           { return nil }
func (alwaysReadyClient) SendProfiles(_ []autoprofile.Profile) error { return nil }
func (alwaysReadyClient) Flush(context.Context) error                { return nil }
