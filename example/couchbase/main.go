// (c) Copyright IBM Corp. 2023

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/couchbase/gocb/v2"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instagocb"

	"net/http"
)

var (
	collector instana.TracerLogger

	// Update this to your cluster details
	connectionString string = "localhost"
	bucketName       string = "travel-sample"
	username         string = "Administrator"
	password         string = "password"
	scopeName        string = "tenant_agent_00"
	collectionName   string = "users"
)

var cluster instagocb.Cluster

func init() {
	ctx := context.Background()
	collector = instana.InitCollector(&instana.Options{
		Service:           "sample-app-couchbase",
		EnableAutoProfile: true,
		Tracer:            instana.DefaultTracerOptions(),
	})

	// init data in couchbase
	dsn := "couchbase://localhost"
	var err error
	cluster, err = instagocb.Connect(collector, dsn, gocb.ClusterOptions{
		Authenticator: gocb.PasswordAuthenticator{
			Username: username,
			Password: password,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	bucketMgr := cluster.Buckets()

	bs := gocb.BucketSettings{
		Name:                 bucketName,
		FlushEnabled:         true,
		ReplicaIndexDisabled: true,
		RAMQuotaMB:           150,
		NumReplicas:          0,
		BucketType:           gocb.CouchbaseBucketType,
	}

	// create bucket
	err = bucketMgr.CreateBucket(gocb.CreateBucketSettings{
		BucketSettings:         bs,
		ConflictResolutionType: gocb.ConflictResolutionTypeSequenceNumber,
	}, &gocb.CreateBucketOptions{
		Context:    ctx,
		ParentSpan: instagocb.GetParentSpanFromContext(ctx),
	})

	if err != nil {
		if !strings.Contains(err.Error(), "Bucket with given name already exists") {
			fmt.Println("Error creating bucket:", err)
			log.Fatal(err)
		} else {
			fmt.Println("Bucket already exists")
			return
		}
	} else {
		fmt.Println("Bucket created successfully")
	}

	bucket := cluster.Bucket(bucketName)

	err = bucket.WaitUntilReady(15*time.Second, nil)
	if err != nil {
		log.Fatal(err)
	}

	// create scope and collection
	collections := bucket.Collections()
	err = collections.CreateScope(scopeName, &gocb.CreateScopeOptions{})
	if err != nil {
		log.Fatal(err)
	}
	err = collections.CreateCollection(gocb.CollectionSpec{
		Name:      collectionName,
		ScopeName: scopeName,
	}, &gocb.CreateCollectionOptions{})
	if err != nil {
		log.Fatal(err)
	}

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
		fmt.Println(err)
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

	bucket := cluster.Bucket(bucketName)

	err := bucket.WaitUntilReady(5*time.Second, nil)
	if err != nil {
		return err
	}

	scope := bucket.Scope("tenant_agent_00")
	col := scope.Collection("users")

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
	var queryStr string

	if needError {
		// invalid_collection is not there, this will produce an error
		queryStr = "SELECT count(*) FROM `" + bucketName + "`." + scopeName + "." + "invalid_collection" + ";"
	} else {
		queryStr = "SELECT count(*) FROM `" + bucketName + "`." + scopeName + "." + collectionName + ";"
	}
	queryResult, err := scope.Query(queryStr,
		&gocb.QueryOptions{
			ParentSpan: instagocb.GetParentSpanFromContext(ctx),
		},
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

	// Starting transactions
	transactions := cluster.Transactions()
	_, err = transactions.Run(func(tac *gocb.TransactionAttemptContext) error {

		// Create new TransactionAttemptContext from instagocb
		tacNew := cluster.WrapTransactionAttemptContext(tac, instagocb.GetParentSpanFromContext(ctx))

		// Unwrapped collection is required to pass it to transaction operations
		collectionUnwrapped := col.Unwrap()

		// Getting documents:
		docA, err := tacNew.Get(collectionUnwrapped, "u:jade")
		// Use err != nil && !errors.Is(err, gocb.ErrDocumentNotFound) if the document may or may not exist
		if err != nil {
			return err
		}

		// Replacing a doc:
		content := User{
			Name:      "Jade New",
			Email:     "jadeNew@test-email.com",
			Interests: []string{"Swimming"},
		}
		err = docA.Content(&content)
		if err != nil {
			return err
		}
		_, err = tacNew.Replace(collectionUnwrapped, docA, content)
		if err != nil {
			return err
		}

		// Removing a doc:
		docA1, err := tacNew.Get(collectionUnwrapped, "u:jade")
		if err != nil {
			return err
		}
		err = tacNew.Remove(collectionUnwrapped, docA1)
		if err != nil {
			return err
		}

		// Performing a SELECT N1QL query against a scope:
		queryResult, err := tacNew.Query("SELECT count(*) FROM `"+bucketName+"`."+scopeName+"."+collectionName+";", &gocb.TransactionQueryOptions{
			// Unwrapped scope is required here
			Scope: scope.Unwrap(),
		})
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
			fmt.Println("Query result : ", result)
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
