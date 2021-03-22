package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//Data Model
type Person struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Firstname string             `json:"firstname" bson:"firstname"`
	Lastname  string             `json:"lastname" bson:"lastname"`
}

var client *mongo.Client

func CreatePersonEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "applications/json")
	var person Person
	json.NewDecoder(request.Body).Decode(&person)
	collection := client.Database("hb").Collection("people")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	result, _ := collection.InsertOne(ctx, person)
	json.NewEncoder(response).Encode(result)
}

func GetPeopleEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "applications/json")

	var people []Person //slice of Person..
	collection := client.Database("hb").Collection("people")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message":"` + err.Error() + `"}`))
		return
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var person Person
		cursor.Decode(&person)
		people = append(people, person)
	}
	if err := cursor.Err(); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message":"` + err.Error() + `"}`))
		return
	}
	json.NewEncoder(response).Encode(people)
}

func GetPersonEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "applications/json")
	params := mux.Vars(request)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	var person Person
	collection := client.Database("hb").Collection("people")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	collection.FindOne(ctx, Person{ID: id}).Decode(&person)
	/*if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message":"` + err.Error() + `"}`))
		return
	}*/
	json.NewEncoder(response).Encode(person)
}

func UpdatePersonEndpoint(response http.ResponseWriter, request *http.Request) {
	// first Delete..
	response.Header().Set("content type", "application/json")
	params := mux.Vars(request)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	collection := client.Database("hb").Collection("people")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	result, err := collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		if err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{"message":"` + err.Error() + `"}`))
			return
		}
		fmt.Println(result)
		json.NewEncoder(response).Encode(result)
	}
	// Now create..

	response.Header().Set("content-type", "applications/json")
	var person Person
	person.ID = id
	json.NewDecoder(request.Body).Decode(&person)
	//collection := client.Database("hb").Collection("people")
	//ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	collection.InsertOne(ctx, person)
	json.NewEncoder(response).Encode(person)

}

func DeletePersonEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content type", "application/json")
	params := mux.Vars(request)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	collection := client.Database("hb").Collection("people")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	result, err := collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		if err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{"message":"` + err.Error() + `"}`))
			return
		}
		fmt.Println(result)
		json.NewEncoder(response).Encode(result)

	}
}
func main() {

	fmt.Println("Starting the application...")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, _ = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))

	router := mux.NewRouter()

	router.HandleFunc("/person", CreatePersonEndpoint).Methods("POST")
	router.HandleFunc("/people", GetPeopleEndpoint).Methods("GET")
	router.HandleFunc("/person/{id}", GetPersonEndpoint).Methods("GET")
	router.HandleFunc("/person/{id}", UpdatePersonEndpoint).Methods("PUT")
	router.HandleFunc("/person/{id}", DeletePersonEndpoint).Methods("DELETE")

	http.ListenAndServe(":12345", router)
}
