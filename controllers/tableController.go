package controllers

import (
	"github.com/menyasosali/restaurant-manage-backend-go/database"
	"net/http"
)

var tableCollection = database.OpenCollection(database.Client, "table")

func GetTables(w http.ResponseWriter, r *http.Request) {

}

func GetTable(w http.ResponseWriter, r *http.Request) {

}

func CreateTable(w http.ResponseWriter, r *http.Request) {

}

func UpdateTable(w http.ResponseWriter, r *http.Request) {

}
