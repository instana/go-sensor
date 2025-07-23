// (c) Copyright IBM Corp. 2025

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	instana "github.com/instana/go-sensor"
	instamongo "github.com/instana/go-sensor/instrumentation/instamongo/v2"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Item defines the data model for an item in the collection.
type Item struct {
	ID   bson.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name string        `json:"name,omitempty" bson:"name,omitempty"`
}

var client *mongo.Client

func agentReady() chan bool {
	ch := make(chan bool)

	go func() {
		for {
			if instana.Ready() {
				ch <- true
			}
		}
	}()

	return ch
}

// connectDB initializes the connection to MongoDB.
func connectDB(sensor instana.TracerLogger) {
	var err error
	client, err = instamongo.Connect(sensor, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal("Error connecting to MongoDB:", err)
	}
}

// createUniqueIndex creates a unique index on the "name" field.
func createUniqueIndex() {
	collection := client.Database("exampledb").Collection("items")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err := collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		log.Printf("Warning: Could not create unique index: %v", err)
		return
	}
	log.Println("Unique index on 'name' created.")
}

// insertHandler handles POST requests to insert a new item.
func insertHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed, only POST is allowed"})
		return
	}

	var item Item
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request payload: " + err.Error()})
		return
	}

	collection := client.Database("exampledb").Collection("items")

	result, err := collection.InsertOne(r.Context(), item)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error inserting item: " + err.Error()})
		return
	}

	response := map[string]interface{}{
		"insertedID": result.InsertedID,
		"message":    "Item inserted successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// getHandler handles GET requests to retrieve all items.
func getHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed, only GET is allowed"})
		return
	}

	collection := client.Database("exampledb").Collection("items")
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	defer cursor.Close(ctx)

	var items []Item
	if err = cursor.All(ctx, &items); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

// errorHandler simulates an insert error by inserting a duplicate item.
func errorHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed, only POST is allowed", http.StatusMethodNotAllowed)
		return
	}

	collection := client.Database("exampledb").Collection("items")

	// Create a sample item with a fixed name.
	item := Item{
		Name: "duplicate-item",
	}
	item.ID = bson.NewObjectID()

	// First insertion should succeed.
	_, err := collection.InsertOne(r.Context(), item)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Initial insert error: " + err.Error()})
		return
	}

	// Second insertion with the same "name" will trigger a duplicate key error.
	item.ID = bson.NewObjectID() // New ID, but same name.
	_, err = collection.InsertOne(r.Context(), item)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Duplicate key error: " + err.Error()})
		return
	}

	// If no error occurred (unexpected), return a success message.
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "No error occurred, but a duplicate key error was expected."})
}

func main() {

	// Initialize Instana collector
	c := instana.InitCollector(&instana.Options{
		Service: "mongo-client",
		Tracer:  instana.DefaultTracerOptions(),
	})
	defer instana.ShutdownCollector()

	<-agentReady()
	log.Print("Agent ready")

	// Connect to MongoDB.
	connectDB(c)
	// create unique index on the collection
	createUniqueIndex()

	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Fatal("Error disconnecting from MongoDB:", err)
		}
	}()

	// Set up HTTP endpoints.
	http.HandleFunc("/insert", instana.TracingHandlerFunc(c, "/insert", insertHandler))
	http.HandleFunc("/get", instana.TracingHandlerFunc(c, "/get", getHandler))
	http.HandleFunc("/error", instana.TracingHandlerFunc(c, "/error", errorHandler))

	fmt.Println("Server is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
