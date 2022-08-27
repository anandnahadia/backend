package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"reflect"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type input struct {
	Data []interface{} `json:"data"`
}

var col *mongo.Collection

func close(client *mongo.Client, ctx context.Context,
	cancel context.CancelFunc) {
	defer cancel()
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
}

func ping(client *mongo.Client, ctx context.Context) error {
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return err
	}
	log.Println("connected successfully")
	return nil
}

// initMongoDb starts mongodb connection
func initMongoDb() (*mongo.Client, context.Context, context.CancelFunc) {
	//start mongodb connection
	clientOptions := options.Client().
		ApplyURI("mongodb://localhost:8082")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	client, err := mongo.Connect(ctx, clientOptions)
	// client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		panic(err)
	}
	col = client.Database("AppBackendData").Collection("appData")
	log.Println("Collection type:", reflect.TypeOf(col))

	// Ping mongoDB with Ping method
	ping(client, ctx)
	return client, ctx, cancel
}
func main() {
	client, ctx, cancel := initMongoDb()
	defer close(client, ctx, cancel)
	r := mux.NewRouter()
	r.HandleFunc("/", helloBackend).Methods("GET")
	r.HandleFunc("/getData", getData).Methods("GET")
	r.HandleFunc("/saveData", addData).Methods("POST")
	log.Println("Listening on 8080........")
	http.ListenAndServe(":8080", r)

}
func helloBackend(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Listening on 8080!"))
}
func addData(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var t input
	err := decoder.Decode(&t)
	if err != nil {
		panic(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 20*time.Second)
	result, insertErr := col.InsertMany(ctx, t.Data)
	if insertErr != nil {
		log.Println("ERROR:", insertErr)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(insertErr)
		return
	}
	log.Println("Data saved successfully ", result)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("success"))
}
func getData(w http.ResponseWriter, r *http.Request) {
	var result interface{}
	err := col.FindOne(context.TODO(), bson.D{}).Decode(&result)
	if err != nil {
		log.Println("ERROR:", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}
	log.Println("get data successfully ", result)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("success"))

}
