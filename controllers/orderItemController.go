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
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	matchStage := bson.D{{"$match", bson.D{{"order_id", id}}}}
	lookupStage := bson.D{{"lookup", bson.D{
		{"from", "food"},
		{"localField", "food_id"},
		{"foreignField", "food_id"},
		{"as", "food"}}}}
	unwindStage := bson.D{{"$unwind", bson.D{{"path", "$food"}, {"preserveNullAndEmptyArrays", true}}}}

	lookupOrderStage := bson.D{{"$lookup", bson.D{
		{"from", "order"},
		{"localField", "order_id"},
		{"foreignField", "order_id"},
		{"as", "food"}}}}
	unwindOrderStage := bson.D{{"$unwind", bson.D{{"path", "$order"}, {"preserveNullAndEmptyArrays", true}}}}

	lookupTableStage := bson.D{{"$lookup", bson.D{
		{"from", "table"},
		{"localField", "order.table_id"},
		{"foreignField", "table_id"},
		{"as", "table"}}}}
	unwindTableStage := bson.D{{"$unwind", bson.D{{"path", "$table"}, {"preserveNullAndEmptyArrays", true}}}}

	projectStage := bson.D{
		{"$project", bson.D{
			{"id", 0},
			{"amount", "$food.price"},
			{"total_count", 1},
			{"food_name", "$food.name"},
			{"food_image", "$food.food_image"},
			{"table_number", "$table.table_number"},
			{"table_id", "$table.table_id"},
			{"order_id", "$order.order_id"},
			{"price", "$food.price"},
			{"quantity", 1},
		}},
	}

	groupStage := bson.D{
		{"$group", bson.D{{"_id", bson.D{
			{"order_id", "$order_id"},
			{"table_id", "$table_id"},
			{"table_number", "$table_number"}}},

			{"payment_due", bson.D{{"$sum", "$amount"},
				{"total_count", bson.D{{"$sum", 1},
					{"order_items", bson.D{{"", ""}}}}}}}}}}

	projectStage2 := bson.D{
		{"$project", bson.D{
			{"id", 0},
			{"payment_due", 1},
			{"total_count", 1},
			{"table_number", "$_id.table_number"},
			{"order_items", 1},
		}},
	}

	result, err := orderItemCollection.Aggregate(ctx, mongo.Pipeline{
		matchStage,
		lookupStage,
		unwindStage,
		lookupOrderStage,
		unwindOrderStage,
		lookupTableStage,
		unwindTableStage,
		projectStage,
		groupStage,
		projectStage2,
	})

	if err != nil {
		panic(err)
	}

	if err := result.All(ctx, &OrderItems); err != nil {
		panic(err)
	}

	return OrderItems, err
}

func CreateOrderItem(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var order models.Order
	var orderItemPack OrderItemPack

	if err := json.NewDecoder(r.Body).Decode(&orderItemPack); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	order.OrderDate, _ = time.Parse(time.RFC822, time.Now().Format(time.RFC822))

	orderItemsToBeInserted := []interface{}{}
	order.TableId = orderItemPack.TableId
	orderId := OrderItemOrderCreator(order)

	for _, orderItem := range orderItemPack.OrderItems {
		orderItem.OrderId = orderId

		validationErr := validate.Struct(orderItem)
		if validationErr != nil {
			http.Error(w, validationErr.Error(), http.StatusBadRequest)
			return
		}

		orderItem.ID = primitive.NewObjectID()
		orderItem.CreatedAt, _ = time.Parse(time.RFC822, time.Now().Format(time.RFC822))
		orderItem.UpdatedAt, _ = time.Parse(time.RFC822, time.Now().Format(time.RFC822))
		orderItem.OrderItemId = orderItem.ID.Hex()
		var num = toFixed(*orderItem.UnitPrice, 2)
		orderItem.UnitPrice = &num
		orderItemsToBeInserted = append(orderItemsToBeInserted, orderItem)
	}

	insertedOrderItems, err := orderItemCollection.InsertMany(ctx, orderItemsToBeInserted)
	if err != nil {
		log.Fatal(err)
	}

	insertedOrderItemsJSON, err := json.Marshal(insertedOrderItems)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(insertedOrderItemsJSON)
}

func UpdateOrderItem(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var orderItem models.OrderItem

	vars := mux.Vars(r)
	orderItemID := vars["order_item_id"]

	if err := json.NewDecoder(r.Body).Decode(&orderItem); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	filter := bson.M{"order_item_id": orderItemID}

	var updateObj primitive.D

	if orderItem.UnitPrice != nil {
		updateObj = append(updateObj, bson.E{"unit_price", *&orderItem.UnitPrice})
	}

	if orderItem.Quantity != nil {
		updateObj = append(updateObj, bson.E{"quantity", *orderItem.Quantity})
	}

	if orderItem.FoodId != nil {
		updateObj = append(updateObj, bson.E{"food_id", *orderItem.FoodId})
	}

	orderItem.UpdatedAt, _ = time.Parse(time.RFC822, time.Now().Format(time.RFC822))
	updateObj = append(updateObj, bson.E{"update_at", orderItem.UpdatedAt})

	upsert := true

	opt := options.UpdateOptions{
		Upsert: &upsert,
	}

	result, err := orderItemCollection.UpdateOne(
		ctx,
		filter,
		bson.D{
			{"$set", updateObj},
		},
		&opt)

	if err != nil {
		msg := "Order items update failed"
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
