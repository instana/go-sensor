package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/couchbase/gocb/v2"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instagocb"

	"encoding/json"
	"io"
	"net/http"

	redis "github.com/redis/go-redis/v9"
)

var (
	c   *http.Client
	rdb *redis.Client
	s   instana.TracerLogger
)

func init() {

	s = instana.InitCollector(&instana.Options{
		Service:           "Nithin-sample-app22",
		EnableAutoProfile: true,
	})

	rdb = redis.NewClient(&redis.Options{Addr: ":6379"})

	c = &http.Client{
		Timeout: time.Second * 30,
	}
}

func main() {
	http.HandleFunc("/star-wars/people", instana.TracingHandlerFunc(s, "/star-wars/people", handler))

	log.Fatal(http.ListenAndServe("localhost:9990", nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	var searchStr []string
	var ok bool
	if searchStr, ok = r.URL.Query()["search"]; !ok {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{message:"please give a search string"}`))
		return
	}

	val, err := rdb.Get(r.Context(), searchStr[0]).Result()
	if err != redis.Nil {
		vb := []byte(val)
		err = gocbMain(r.Context())

		if err == nil {
			sendOkResp(w, vb, false)
		} else {
			sendErrResp(w)
		}
		return

	}

	resp, err := c.Get(`https://swapi.dev/api/people/?search=` + searchStr[0] + `&format=json`)
	if err != nil {
		sendErrResp(w)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		sendErrResp(w)
		return
	}

	rdb.Set(r.Context(), searchStr[0], body, time.Duration(time.Second*10))

	err = gocbMain(r.Context())

	if err == nil {
		sendOkResp(w, body, false)
	} else {
		sendErrResp(w)
	}

}

func sendErrResp(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write([]byte(`{message:"some internal error"}`))
}

func sendOkResp(w http.ResponseWriter, vb []byte, cached bool) {

	res := make(map[string]any)

	err := json.Unmarshal(vb, &res)
	if err != nil {
		sendErrResp(w)
		return
	}

	res["cached"] = cached

	resStr, err := json.Marshal(res)
	if err != nil {
		sendErrResp(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(resStr)
}

func gocbMain(ctx context.Context) error {
	// hold := make(chan bool)
	// Uncomment following line to enable logging
	// gocb.SetLogger(gocb.VerboseStdioLogger())

	// Update this to your cluster details
	connectionString := "localhost"
	bucketName := "travel-sample"
	username := "Administrator"
	password := "password"

	dsn := "couchbase://" + connectionString

	// s := instana.InitCollector(&instana.Options{
	// 	Service: "nithin-couchbase-app1",
	// })

	// For a secure cluster connection, use `couchbases://<your-cluster-ip>` instead.
	cluster, err := gocb.Connect(dsn, gocb.ClusterOptions{
		Authenticator: gocb.PasswordAuthenticator{
			Username: username,
			Password: password,
		},
		Tracer: instagocb.NewTracer(s, dsn),
	})
	if err != nil {
		return err
	}

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

	fmt.Println("holding process up")
	// <-hold

	return nil
}
