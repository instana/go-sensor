// (c) Copyright IBM Corp. 2022

//go:build integration
// +build integration

package instapgx_test

import (
	"context"
	"database/sql"
	"log"
	"math/rand"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instapgx"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
)

var databaseUrl = "postgres://postgres:mysecretpassword@localhost/"

var dbNamePrefix = "db_"

func TestMain(m *testing.M) {

	// Create test db
	databaseUrlOrg := databaseUrl
	db, err := sql.Open("postgres", databaseUrl+"?sslmode=disable")

	if err != nil {
		log.Fatalln("Can not connect to the database:", err.Error())
	}

	dbName := dbNamePrefix + strings.Replace(uuid.New().String(), "-", "", -1)
	_, err = db.Exec("CREATE DATABASE " + dbName)

	if err != nil {
		log.Fatalln("Can not create database: ", err.Error())
	}

	db.Close()

	// Connect to test db
	databaseUrl = databaseUrl + dbName
	db, err = sql.Open("postgres", databaseUrl+"?sslmode=disable")

	if err != nil {
		log.Fatalln("Can not connect to the database:", err.Error())
	}

	// Running tests
	code := m.Run()
	db.Close()

	// Clean test db
	db, err = sql.Open("postgres", databaseUrlOrg+"?sslmode=disable")

	if err != nil {
		log.Fatalln("Can not connect to the database:", err.Error())
	}

	defer db.Close()

	_, err = db.Exec("DROP DATABASE " + dbName + " WITH (FORCE)")

	if err != nil {
		log.Fatalln("Can not drop database: "+dbName, err.Error())
	}

	os.Exit(code)
}

func prepare(t *testing.T) (*instana.Recorder, context.Context, *instapgx.Conn) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	sensor := instana.NewSensorWithTracer(tracer)

	conf, err := pgx.ParseConfig(databaseUrl)
	assert.NoError(t, err)

	pSpan := sensor.Tracer().StartSpan("parent-span")
	ctx := context.Background()
	if pSpan != nil {
		ctx = instana.ContextWithSpan(ctx, pSpan)
	}
	conn, err := instapgx.ConnectConfig(ctx, sensor, conf)

	assert.NoError(t, err)
	assert.IsType(t, &instapgx.Conn{}, conn)
	return recorder, ctx, conn
}

func randStringBytes(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
