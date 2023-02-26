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

var orderCollection = database.OpenCollection(database.Client, "order")

func GetOrders(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	result, err := orderCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		msg := "error occurred while listing order items"
		http.Error(w, msg, http.StatusInternalServerError)
	}

	var allOrders []bson.M
	if err = result.All(ctx, &allOrders); err != nil {
		log.Fatal(err)
	}

	allOrdersJSON, err := json.Marshal(allOrders)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}
	w.WriteHeader(http.StatusOK)
	w.Write(allOrdersJSON)
}

func GetOrder(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	vars := mux.Vars(r)
	orderId := vars["order_id"]

	var order models.Order

	if err := orderCollection.FindOne(ctx, orderId).Decode(&order); err != nil {
		msg := "error occurred while listing orders"
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	orderJSON, err := json.Marshal(order)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(orderJSON)
}

func CreateOrder(w http.ResponseWriter, r *http.Request) {
	var order models.Order
	var table models.Table
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	validationErr := validate.Struct(order)
	if validationErr != nil {
		http.Error(w, validationErr.Error(), http.StatusBadRequest)
		return
	}

	if order.TableId != nil {
		err := tableCollection.FindOne(ctx, bson.M{"table_id": order.TableId}).Decode(&table)
		if err != nil {
			msg := fmt.Sprintf("message: Table was not found")
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}
	}

	order.CreatedAt, _ = time.Parse(time.RFC822, time.Now().Format(time.RFC822))
	order.UpdatedAt, _ = time.Parse(time.RFC822, time.Now().Format(time.RFC822))
	order.ID = primitive.NewObjectID()
	order.OrderId = order.ID.Hex()

	result, insertErr := orderCollection.InsertOne(ctx, order)
	if insertErr != nil {
		msg := fmt.Sprintf("Order item was not created")
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

func UpdateOrder(w http.ResponseWriter, r *http.Request) {
	var order models.Order
	var table models.Table
	var updateObj primitive.D

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	vars := mux.Vars(r)
	orderId := vars["order_id"]

	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if order.TableId != nil {
		if err := orderCollection.FindOne(ctx, bson.M{"order_id": table.TableId}).Decode(&table); err != nil {
			msg := fmt.Sprintf("message: Order was not found")
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}
		updateObj = append(updateObj, bson.E{"menu", order.TableId})
	}

	order.UpdatedAt, _ = time.Parse(time.RFC822, time.Now().Format(time.RFC822))
	updateObj = append(updateObj, bson.E{"updated_at", order.UpdatedAt})

	upsert := true

	opt := options.UpdateOptions{
		Upsert: &upsert,
	}

	filter := bson.M{"order_id": orderId}

	result, err := orderCollection.UpdateOne(
		ctx,
		filter,
		bson.D{
			{"$set", updateObj},
		},
		&opt,
	)
	if err != nil {
		msg := "Order update failed"
		http.Error(w, msg, http.StatusInternalServerError)
	}

	resultJson, err := json.Marshal(result)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resultJson)
}

func OrderItemOrderCreator(order models.Order) string {
	order.CreatedAt, _ = time.Parse(time.RFC822, time.Now().Format(time.RFC822))
	order.UpdatedAt, _ = time.Parse(time.RFC822, time.Now().Format(time.RFC822))
	order.ID = primitive.NewObjectID()
	order.OrderId = order.ID.Hex()

	orderCollection.InsertOne(ctx, order)

	return order.OrderId
}
