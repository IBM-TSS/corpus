package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Ticket struct {
	ID  string `json:"_id,omitempty" bson:"_id,omitempty"`
	Esn string `json:"esn,omitempty" bson:"esn,omitempty"`
}

var client *mongo.Client
var long_run_client *mongo.Client
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

func GetUnAssignedTicket(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")
	collection := client.Database("tickets").Collection("not_assigned")
	bdi_collection := long_run_client.Database("santander").Collection("bdi")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()
	
	var result []bson.M
	var err error
	var cursor *mongo.Cursor
	
	cursor, err = collection.Find(ctx, bson.D{})
	defer cursor.Close(ctx)

	if err != nil {
		// ErrNoDocuments means that the filter did not match any documents in the collection
		if err == mongo.ErrNoDocuments {
			log.Fatal("noo doocss")
			return
		}
		log.Fatal("fui")
		log.Fatal(err)
	}

	for cursor.Next(ctx) {
		var temp bson.M
		var atm bson.M

		cursor.Decode(&temp)
		// Find the ATM
		err = bdi_collection.FindOne(ctx, bson.D{{"_id", temp["atm"]}}).Decode(&atm)
		if err != nil {
			// ErrNoDocuments means that the filter did not match any documents in the collection
			if err == mongo.ErrNoDocuments {
				log.Fatal("noo doocss")
				return
			}
			log.Fatal(temp)
			log.Fatal(err)
		}

		temp["atm"] = atm
		result = append(result, temp)

	}

	json.NewEncoder(response).Encode(result)
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
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

	long_run_client, err = mongo.Connect(ctx, options.Client().ApplyURI(
		"mongodb+srv://eduardoSantos:05120714@cluster0.mx9ah.mongodb.net/santander?retryWrites=true&w=majority",
	))

	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()

	
	router.HandleFunc("/ticket", CreateTicketEndpoint).Methods("POST")
	// router.HandleFunc("/ticket/{_id}", GetUnAssignedTicket).Methods("GET")
	router.HandleFunc("/ticket", GetUnAssignedTicket).Methods("GET")

	handler := cors.AllowAll().Handler(router)
	log.Fatal(http.ListenAndServe(":9707", handler))
}
