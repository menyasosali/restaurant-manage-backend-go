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
	"log"
	"net/http"
	"time"
)

type OrderItemPack struct {
	TableId    *string
	OrderItems []models.OrderItem
}

var orderItemCollection = database.OpenCollection(database.Client, "orderItem")

func GetOrderItems(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	results, err := orderItemCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		msg := "error occurred while listing order items"
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	var allOrderItems []bson.M
	if err = results.All(ctx, &allOrderItems); err != nil {
		log.Fatal(err)
	}

	allOrderItemsJSON, err := json.Marshal(allOrderItems)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}
	w.WriteHeader(http.StatusOK)
	w.Write(allOrderItemsJSON)
}

func GetOrderItem(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var orderItem models.OrderItem

	vars := mux.Vars(r)
	orderItemId := vars["order_item_id"]

	if err := orderItemCollection.FindOne(ctx, bson.M{"order_item_id": orderItemId}).Decode(&orderItem); err != nil {
		msg := fmt.Sprintf("error occurred while listing orders")
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	orderItemJSON, err := json.Marshal(orderItem)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}
	w.WriteHeader(http.StatusOK)
	w.Write(orderItemJSON)
}

func GetOrderItemsByOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderId := vars["order_id"]

	allOrderItems, err := ItemsByOrder(orderId)
	if err != nil {
		msg := fmt.Sprintf("error occurred while listing order items by order ID")
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	allOrderItemsJSON, err := json.Marshal(allOrderItems)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(allOrderItemsJSON)
}

func ItemsByOrder(id string) (OrderItems []primitive.M, err error) {

}

func CreateOrderItem(w http.ResponseWriter, r *http.Request) {

}

func UpdateOrderItem(w http.ResponseWriter, r *http.Request) {

}
