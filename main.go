package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Ticket struct {
	ID  string `json:"_id,omitempty" bson:"_id,omitempty"`
	Esn string `json:"esn,omitempty" bson:"esn,omitempty"`
}

var client *mongo.Client
var err error

func CreateTicketEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")

	var ticket Ticket
	json.NewDecoder(request.Body).Decode(&ticket)

	collection := client.Database("tickets").Collection("not_assigned")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	result, _ := collection.InsertOne(ctx, ticket)
	json.NewEncoder(response).Encode(result)
}

func main() {
	fmt.Println("Starting")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(
		"mongodb+srv://eduardo:05120714@cluster0.9byad.mongodb.net/tickets?retryWrites=true&w=majority",
	))

	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()

	router.HandleFunc("/ticket", CreateTicketEndpoint).Methods("POST")

	http.ListenAndServe(":9707", router)

}
