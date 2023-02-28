package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/menyasosali/restaurant-manage-backend-go/database"
	"github.com/menyasosali/restaurant-manage-backend-go/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"time"
)

var tableCollection = database.OpenCollection(database.Client, "table")

func GetTables(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	result, err := tableCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		msg := "error occurred while listing tables"
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	var allTables []bson.M

	if err = result.All(ctx, &allTables); err != nil {
		log.Fatal(err)
	}

	allTablesJSON, err := json.Marshal(allTables)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(allTablesJSON)
}

func GetTable(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var table models.Table
	vars := mux.Vars(r)
	tableID := vars["table_id"]

	if err := tableCollection.FindOne(ctx, bson.M{"table_id": tableID}).Decode(&table); err != nil {
		msg := fmt.Sprintf("error occurred while listing tables")
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	tableJSON, err := json.Marshal(table)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(tableJSON)

}

func CreateTable(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var table models.Table

	if err := json.NewDecoder(r.Body).Decode(&table); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	validationErr := validate.Struct(table)
	if validationErr != nil {
		http.Error(w, validationErr.Error(), http.StatusBadRequest)
		return
	}

	table.ID = primitive.NewObjectID()
	table.CreatedAt, _ = time.Parse(time.RFC822, time.Now().Format(time.RFC822))
	table.UpdatedAt, _ = time.Parse(time.RFC822, time.Now().Format(time.RFC822))
	table.TableId = table.ID.Hex()

	result, err := tableCollection.InsertOne(ctx, table)
	if err != nil {
		msg := fmt.Sprintf("Table item was not created")
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resultJSON)
}

func UpdateTable(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var table models.Table

	var updateObj primitive.D

	if err := json.NewDecoder(r.Body).Decode(&table); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	tableID := vars["table_id"]

	if table.TableNumber != nil {
		updateObj = append(updateObj, bson.E{"table_number", table.TableNumber})
	}

	if table.NumberOfGuests != nil {
		updateObj = append(updateObj, bson.E{"number_of_guests", table.NumberOfGuests})
	}

	table.UpdatedAt, _ = time.Parse(time.RFC822, time.Now().Format(time.RFC822))
	updateObj = append(updateObj, bson.E{"updated_at", table.UpdatedAt})

	upsert := true

	opt := options.UpdateOptions{
		Upsert: &upsert,
	}

	filter := bson.E{"table_id", tableID}

	result, err := tableCollection.UpdateOne(
		ctx,
		filter,
		bson.D{
			{"$set", updateObj},
		},
		&opt)

	if err != nil {
		msg := fmt.Sprintf("Table item update failed")
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resultJSON)
}
