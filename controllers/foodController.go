package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/menyasosali/restaurant-manage-backend-go/database"
	"github.com/menyasosali/restaurant-manage-backend-go/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"
)

var foodCollection = database.OpenCollection(database.Client, "food")
var validate = validator.New()

func GetFoods(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	recordPerPage, err := strconv.Atoi(r.FormValue("recordPerPage"))
	if err != nil || recordPerPage < 1 {
		recordPerPage = 10
	}
	page, err := strconv.Atoi(r.FormValue("page"))
	if err != nil || page < 1 {
		page = 1
	}

	startIndex := (page - 1) * recordPerPage
	startIndex, err = strconv.Atoi(r.FormValue("startIndex"))

	matchStage := bson.D{{"match", bson.D{}}}
	groupStage := bson.D{{"group", bson.D{{"_id", bson.D{{"_id", "null"}}},
		{"total_count", bson.D{{"$sum", 1}}},
		{"data", bson.D{{"$push", "$$ROOT"}}},
	}}}
	projectStage := bson.D{
		{
			"$project", bson.D{
				{"_id", 0},
				{"total_count", 1},
				{"food_items", bson.D{{"$slice", []interface{}{"$data", startIndex, recordPerPage}}}},
			},
		},
	}

	result, err := foodCollection.Aggregate(ctx, mongo.Pipeline{
		matchStage, groupStage, projectStage,
	})
	if err != nil {
		msg := "error occurred while listing food items"
		http.Error(w, msg, http.StatusInternalServerError)
	}
	var allFoods []bson.M
	if err = result.All(ctx, &allFoods); err != nil {
		log.Fatal(err)
	}

	allFoodsJSON, err := json.Marshal(allFoods[0])
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}
	w.WriteHeader(http.StatusOK)
	w.Write(allFoodsJSON)

}

func GetFood(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	vars := mux.Vars(r)
	foodId := vars["food_id"]
	var food models.Food

	err := foodCollection.FindOne(ctx, bson.M{"food_id": foodId}).Decode(&food)
	if err != nil {
		http.Error(w, "error occurred while fetching the food item", http.StatusInternalServerError)
	}

	foodJSON, err := json.Marshal(food)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(foodJSON)
}

func CreateFood(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	var menu models.Menu
	var food models.Food

	if err := json.NewDecoder(r.Body).Decode(&food); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	validationErr := validate.Struct(food)
	if validationErr != nil {
		http.Error(w, validationErr.Error(), http.StatusBadRequest)
		return
	}

	if err := menuCollection.FindOne(ctx, bson.M{"menu_id": menu.MenuId}).Decode(&menu); err != nil {
		msg := fmt.Sprintf("menu was not found")
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	food.CreatedAt, _ = time.Parse(time.RFC822, time.Now().Format(time.RFC822))
	food.UpdatedAt, _ = time.Parse(time.RFC822, time.Now().Format(time.RFC822))
	food.ID = primitive.NewObjectID()
	food.FoodId = food.ID.Hex()
	var num = toFixed(*food.Price, 2)
	food.Price = &num

	result, insertErr := foodCollection.InsertOne(ctx, food)
	if insertErr != nil {
		msg := fmt.Sprintf("Food item was not created")
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	resJSON, err := json.Marshal(result)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resJSON)
}

func UpdateFood(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	var food models.Food
	var menu models.Menu

	if err := json.NewDecoder(r.Body).Decode(&food); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)

	foodId := vars["food_id"]
	filter := bson.M{"food_id": foodId}

	var updateObj primitive.D

	if food.Name != nil {
		updateObj = append(updateObj, bson.E{"name", food.Name})
	}

	if food.Price != nil {
		updateObj = append(updateObj, bson.E{"price", food.Price})
	}

	if food.FoodImage != nil {
		updateObj = append(updateObj, bson.E{"food_image", food.FoodImage})
	}

	if food.MenuId != nil {
		err := menuCollection.FindOne(ctx, bson.M{"menu_id": food.MenuId}).Decode(&menu)
		if err != nil {
			msg := "message: Menu was not found"
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}
		updateObj = append(updateObj, bson.E{"menu", food.MenuId})
	}

	food.UpdatedAt, _ = time.Parse(time.RFC822, time.Now().Format(time.RFC822))
	updateObj = append(updateObj, bson.E{"update_at", food.UpdatedAt})

	upset := true

	opt := options.UpdateOptions{
		Upsert: &upset,
	}

	result, err := foodCollection.UpdateOne(
		ctx,
		filter,
		bson.D{
			{"$set", updateObj},
		},
		&opt,
	)
	if err != nil {
		msg := "Food update failed"
		http.Error(w, msg, http.StatusInternalServerError)
	}

	resultJson, err := json.Marshal(result)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resultJson)
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}
