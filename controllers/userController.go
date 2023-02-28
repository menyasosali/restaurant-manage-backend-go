package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/menyasosali/restaurant-manage-backend-go/database"
	"github.com/menyasosali/restaurant-manage-backend-go/helpers"
	"github.com/menyasosali/restaurant-manage-backend-go/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"strconv"
	"time"
)

var userCollection = database.OpenCollection(database.Client, "user")

func GetUsers(w http.ResponseWriter, r *http.Request) {
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

	matchStage := bson.D{{"$match", bson.D{{}}}}
	projectStage := bson.D{
		{
			"$project", bson.D{
				{"_id", "0"},
				{"total_count", "1"},
				{"user_items", bson.D{{"$slice", []interface{}{"$data", startIndex, recordPerPage}}}},
			},
		},
	}

	result, err := userCollection.Aggregate(ctx, mongo.Pipeline{
		matchStage, projectStage,
	})
	if err != nil {
		msg := "error occurred while listing user items"
		http.Error(w, msg, http.StatusInternalServerError)
	}
	var allUsers []bson.M
	if err := result.All(ctx, &allUsers); err != nil {
		log.Fatal(err)
		return
	}

	allUsersJSON, err := json.Marshal(allUsers)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(allUsersJSON)
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var user models.User

	vars := mux.Vars(r)
	userId := vars["user_id"]

	err := userCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user)

	if err != nil {
		msg := "error occurred while listing users"
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	userJSON, err := json.Marshal(user)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(userJSON)
}

func SingUp(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var user models.User

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	validationErr := validate.Struct(user)
	if validationErr != nil {
		http.Error(w, validationErr.Error(), http.StatusBadRequest)
		return
	}

	count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
	if err != nil {
		msg := "error occurred while checking for the email"
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	password := HashPassword(*user.Password)
	user.Password = &password

	count, err = userCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
	if err != nil {
		msg := "error occurred while checking for the phone number"
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	if count > 0 {
		msg := "this email or phone number already exists"
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	user.CreatedAt, _ = time.Parse(time.RFC822, time.Now().Format(time.RFC822))
	user.UpdatedAt, _ = time.Parse(time.RFC822, time.Now().Format(time.RFC822))
	user.ID = primitive.NewObjectID()
	user.UserId = user.ID.Hex()

	token, refreshToken, _ := helpers.GenerateAllTokens(*user.Email, *user.FirstName, *user.SecondName, user.UserId)
	user.Token = &token
	user.RefreshToken = &refreshToken

	resultInsertNumber, insertErr := userCollection.InsertOne(ctx, user)
	if insertErr != nil {
		msg := fmt.Sprintf("User item was not created")
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	resultInsertNumberJSON, err := json.Marshal(resultInsertNumber)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resultInsertNumberJSON)
}

func Login(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var user models.User
	var foundUser models.User

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
	if err != nil {
		msg := "user not found, login seems to be incorrect"
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	passwordValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
	if passwordValid != true {
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	token, refreshToken, _ := helpers.GenerateAllTokens(*foundUser.Email, *foundUser.FirstName, *foundUser.SecondName, foundUser.UserId)

	helpers.UpdateAllTokens(token, refreshToken, foundUser.UserId)

	foundUserJSON, err := json.Marshal(foundUser)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(foundUserJSON)
}

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		panic(err)
	}
	return string(bytes)
}

func VerifyPassword(userPassword string, providedPassword string) (bool, string) {

	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true
	msg := ""

	if err != nil {
		msg = fmt.Sprintf("login or password is incorrect")
		check = false
	}
	return check, msg
}
