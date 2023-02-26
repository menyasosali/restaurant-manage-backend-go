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

var menuCollection = database.OpenCollection(database.Client, "menu")

func GetMenus(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	result, err := menuCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		http.Error(w, "error occurred while listing the menu item", http.StatusBadRequest)
	}

	var allMenus []bson.M
	if err = result.All(ctx, &allMenus); err != nil {
		log.Fatal(err)
	}

	allMenusJSON, err := json.Marshal(allMenus)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(allMenusJSON)
}

func GetMenu(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	vars := mux.Vars(r)
	menuId := vars["menu_id"]

	var menu models.Menu

	err := menuCollection.FindOne(ctx, bson.M{"menu_id": menuId}).Decode(&menu)
	if err != nil {
		http.Error(w, "occurred while fetching the menu", http.StatusInternalServerError)
	}

	menuJSON, err := json.Marshal(menu)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(menuJSON)
}

func CreateMenu(w http.ResponseWriter, r *http.Request) {
	var menu models.Menu
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	if err := json.NewDecoder(r.Body).Decode(&menu); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	validationErr := validate.Struct(menu)
	if validationErr != nil {
		http.Error(w, validationErr.Error(), http.StatusBadRequest)
		return
	}

	menu.CreatedAt, _ = time.Parse(time.RFC822, time.Now().Format(time.RFC822))
	menu.UpdatedAt, _ = time.Parse(time.RFC822, time.Now().Format(time.RFC822))
	menu.ID = primitive.NewObjectID()
	menu.MenuId = menu.ID.Hex()

	result, insertErr := menuCollection.InsertOne(ctx, menu)
	if insertErr != nil {
		msg := fmt.Sprintf("Menu item was not created")
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

func UpdateMenu(w http.ResponseWriter, r *http.Request) {
	var menu models.Menu
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	if err := json.NewDecoder(r.Body).Decode(&menu); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	vars := mux.Vars(r)

	menuId := vars["menu_id"]
	filter := bson.M{"menu_id": menuId}

	var updateObj primitive.D

	if menu.StartDate != nil && menu.EndDate != nil {
		if !inTimeSpan(*menu.StartDate, *menu.EndDate, time.Now()) {
			msg := "kindly retype the time"
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}

		updateObj = append(updateObj, bson.E{"start_date", menu.StartDate})
		updateObj = append(updateObj, bson.E{"end_date", menu.EndDate})

		if menu.Name != "" {
			updateObj = append(updateObj, bson.E{"name", menu.Name})
		}

		if menu.Category != "" {
			updateObj = append(updateObj, bson.E{"category", menu.Category})
		}

		menu.UpdatedAt, _ = time.Parse(time.RFC822, time.Now().Format(time.RFC822))
		updateObj = append(updateObj, bson.E{"update_at", menu.UpdatedAt})

		upsert := true

		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		result, err := menuCollection.UpdateOne(
			ctx,
			filter,
			bson.D{
				{"$set", updateObj},
			},
			&opt,
		)

		if err != nil {
			msg := "Menu update failed"
			http.Error(w, msg, http.StatusInternalServerError)
		}

		resultJson, err := json.Marshal(result)
		if err != nil {
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
		}

		w.WriteHeader(http.StatusOK)
		w.Write(resultJson)
	}
}

func inTimeSpan(start, end, check time.Time) bool {
	return start.After(check) && end.After(start)
}
