// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package instamongo_test

import (
	"context"
	"fmt"
	"log"
	"time"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instamongo/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const localhostMongo = "mongodb://localhost:27017"

// The following example demonstrates how to instrument a MongoDB client with instana using
// github.com/instana/go-sensor/instrumentation/instamongo wrapper module.
func Example() {
	// Initialize Instana collector
	c := instana.InitCollector(&instana.Options{
		Service: "mongo-client",
	})
	defer instana.ShutdownCollector()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use instamongo.Connect() to establish connection to MongoDB and instrument the client
	client, err := instamongo.Connect(c, options.Client().ApplyURI(localhostMongo))
	if err != nil {
		log.Fatalf("failed to connect to %s: %s", localhostMongo, err)
	}

	// Use instrumented client as usual
	dbs, err := client.ListDatabases(ctx, bson.D{})
	if err != nil {
		log.Fatalf("failed to list databases: %s", err)
	}

	fmt.Println("found", len(dbs.Databases), "database(s)")
	for _, db := range dbs.Databases {
		fmt.Println("* ", db.Name)
	}
}
