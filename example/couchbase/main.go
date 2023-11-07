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
	c *http.Client
	s instana.TracerLogger
)

func init() {

	s = instana.InitCollector(&instana.Options{
		Service:           "Nithin-sample-app-couchbase",
		EnableAutoProfile: true,
	})

	c = &http.Client{
		Timeout: time.Second * 30,
	}
}

func main() {
	http.HandleFunc("/couchbase-test", instana.TracingHandlerFunc(s, "/couchbase-test", handler))

	log.Fatal(http.ListenAndServe("localhost:9990", nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	err := testCouchbase(r.Context())

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

func testCouchbase(ctx context.Context) error {

	// Update this to your cluster details
	connectionString := "localhost"
	bucketName := "travel-sample"
	username := "Administrator"
	password := "password"

	dsn := "couchbase://" + connectionString

	// For a secure cluster connection, use `couchbases://<your-cluster-ip>` instead.

	t := instagocb.NewTracer(s, dsn)
	cluster, err := gocb.Connect(dsn, gocb.ClusterOptions{
		Authenticator: gocb.PasswordAuthenticator{
			Username: username,
			Password: password,
		},
		Tracer: t,
	})
	if err != nil {
		return err
	}

	// wrapping the connected cluster in tracer
	t.WrapCluster(cluster)

	bucket := cluster.Bucket(bucketName)

	err = bucket.WaitUntilReady(5*time.Second, nil)
	if err != nil {
		return err
	}

	// Get a reference to the default collection, required for older Couchbase server versions
	// col := bucket.DefaultCollection()

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
		}, &gocb.UpsertOptions{ParentSpan: instagocb.GetParentSpanFromContext(ctx)})
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

	// Perform a N1QL Query
	inventoryScope := bucket.Scope("inventory")
	queryResult, err := inventoryScope.Query(
		fmt.Sprintf("SELECT * FROM airline WHERE id=10"),
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
