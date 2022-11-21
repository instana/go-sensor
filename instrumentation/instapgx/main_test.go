// (c) Copyright IBM Corp. 2022

//go:build integration
// +build integration

package instapgx_test

import (
	"context"
	"database/sql"
	"log"
	"math/rand"
	"testing"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instapgx"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
)

var databaseUrl = "postgres://postgres:mysecretpassword@localhost/postgres"

func TestMain(m *testing.M) {
	db, err := sql.Open("postgres", databaseUrl)
	if err != nil {
		log.Fatalln("Can not connect to the database:", err.Error())
	}
	defer db.Close()

	m.Run()
}

func prepare(t *testing.T) (*instana.Recorder, context.Context, *instapgx.Conn) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	sensor := instana.NewSensorWithTracer(tracer)

	conf, err := pgx.ParseConfig(databaseUrl)
	assert.NoError(t, err)
	ctx := context.Background()
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
