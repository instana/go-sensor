// (c) Copyright IBM Corp. 2023

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/couchbase/gocb/v2"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instagocb"

	"net/http"
)

var (
	collector instana.TracerLogger
)

func init() {
	collector = instana.InitCollector(&instana.Options{
		Service:           "sample-app-couchbase",
		EnableAutoProfile: true,
	})
}

func main() {
	http.HandleFunc("/couchbase-test", instana.TracingHandlerFunc(collector, "/couchbase-test", handler))
	log.Fatal(http.ListenAndServe("localhost:9990", nil))
}

func handler(w http.ResponseWriter, r *http.Request) {

	var needError bool

	erStr := r.URL.Query().Get("error")
	if erStr == "true" {
		needError = true
	}

	err := couchbaseTest(r.Context(), needError)

	if err != nil {
		sendErrResp(w)
		return
	}

	sendOkResp(w)
}

func sendErrResp(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write([]byte(`{message:"some internal error"}`))
}

func sendOkResp(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("{message:Status OK! Check terminal for full log!}"))
}

func couchbaseTest(ctx context.Context, needError bool) error {

	// Update this to your cluster details
	connectionString := "localhost"
	bucketName := "travel-sample"
	username := "Administrator"
	password := "password"

	dsn := "couchbase://" + connectionString

	cluster, err := instagocb.Connect(collector, dsn, gocb.ClusterOptions{
		Authenticator: gocb.PasswordAuthenticator{
			Username: username,
			Password: password,
		},
	})
	if err != nil {
		return err
	}

	bucket := cluster.Bucket(bucketName)

	err = bucket.WaitUntilReady(5*time.Second, nil)
	if err != nil {
		return err
	}

	col := bucket.Scope("tenant_agent_00").Collection("users")

	type User struct {
		Name      string   `json:"name"`
		Email     string   `json:"email"`
		Interests []string `json:"interests"`
	}

	// Create and store a Document
	_, err = col.Upsert("u:jade",
		User{
			Name:      "Jade",
			Email:     "jade@test-email.com",
			Interests: []string{"Swimming", "Rowing"},
		}, &gocb.UpsertOptions{
			// If you are using couchbase call as part of some http request or something,
			// you need to set this parentSpan property using `instagocb.GetParentSpanFromContext` method,
			// Else the parent-child span relationship wont be tracked.
			// You can keep this as nil, otherwise.

			// Here in this case, this couchbase call is part of a http request.
			ParentSpan: instagocb.GetParentSpanFromContext(ctx),
		})
	if err != nil {
		return err
	}

	// Get the document back
	getResult, err := col.Get("u:jade", &gocb.GetOptions{ParentSpan: instagocb.GetParentSpanFromContext(ctx)})
	if err != nil {
		return err
	}

	var inUser User
	err = getResult.Content(&inUser)
	if err != nil {
		return err
	}
	fmt.Printf("User: %v\n", inUser)

	// Perform a SQL++ Query
	inventoryScope := bucket.Scope("inventory")
	var queryStr string

	if needError {
		// airline1 is not there, this will produce an error
		queryStr = "SELECT * FROM `airline1` WHERE id=10"
	} else {
		queryStr = "SELECT * FROM `airline` WHERE id=10"
	}
	queryResult, err := inventoryScope.Query(queryStr,
		&gocb.QueryOptions{Adhoc: true, ParentSpan: instagocb.GetParentSpanFromContext(ctx)},
	)
	if err != nil {
		return err
	}

	// Print each found Row
	for queryResult.Next() {
		var result interface{}
		err := queryResult.Row(&result)
		if err != nil {
			return err
		}
		fmt.Println(result)
	}

	if err := queryResult.Err(); err != nil {
		return err
	}

	return nil
}
