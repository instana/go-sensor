// (c) Copyright IBM Corp. 2023

package main

import (
	"context"
	"errors"
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

	// Test transactions
	collection := inventoryScope.Collection("airport")

	// Starting transactions
	transactions := cluster.Transactions()
	_, err = transactions.Run(func(tac *gocb.TransactionAttemptContext) error {

		// Create new TransactionAttemptContext from instagocb
		tacNew := cluster.WrapTransactionAttemptContext(tac, instagocb.GetParentSpanFromContext(ctx))

		// Unwrapped collection is required to pass it to transaction operations
		collectionUnwrapped := collection.Unwrap()

		// Inserting a doc:
		_, err := tacNew.Insert(collectionUnwrapped, "doc-a", map[string]interface{}{})
		if err != nil {
			return err
		}

		// Getting documents:
		docA, err := tacNew.Get(collectionUnwrapped, "doc-a")
		// Use err != nil && !errors.Is(err, gocb.ErrDocumentNotFound) if the document may or may not exist
		if err != nil {
			return err
		}

		// Replacing a doc:
		var content map[string]interface{}
		err = docA.Content(&content)
		if err != nil {
			return err
		}
		content["transactions"] = "are awesome"
		_, err = tacNew.Replace(collectionUnwrapped, docA, content)
		if err != nil {
			return err
		}

		// Removing a doc:
		docA1, err := tacNew.Get(collectionUnwrapped, "doc-a")
		if err != nil {
			return err
		}
		err = tacNew.Remove(collectionUnwrapped, docA1)
		if err != nil {
			return err
		}

		// Performing a SELECT N1QL query against a scope:
		qr, err := tacNew.Query("SELECT * FROM hotel WHERE country = $1", &gocb.TransactionQueryOptions{
			PositionalParameters: []interface{}{"United Kingdom"},

			// Unwrapped scope is required here
			Scope: inventoryScope.Unwrap(),
		})
		if err != nil {
			return err
		}

		type hotel struct {
			Name string `json:"name"`
		}

		var hotels []hotel
		for qr.Next() {
			var h hotel
			err = qr.Row(&h)
			if err != nil {
				return err
			}

			hotels = append(hotels, h)
		}

		// Performing an UPDATE N1QL query on multiple documents, in the `inventory` scope:
		_, err = tacNew.Query("UPDATE route SET airlineid = $1 WHERE airline = $2", &gocb.TransactionQueryOptions{
			PositionalParameters: []interface{}{"airline_137", "AF"},
			// Unwrapped scope is required here
			Scope: inventoryScope.Unwrap(),
		})
		if err != nil {
			return err
		}

		// There is no commit call, by not returning an error the transaction will automatically commit
		return nil
	}, nil)
	var ambigErr gocb.TransactionCommitAmbiguousError
	if errors.As(err, &ambigErr) {
		log.Println("Transaction possibly committed")

		log.Printf("%+v", ambigErr)
		return nil
	}
	var failedErr gocb.TransactionFailedError
	if errors.As(err, &failedErr) {
		log.Println("Transaction did not reach commit point")

		log.Printf("%+v", failedErr)
		return nil
	}

	if err != nil {
		return err
	}

	return nil
}
