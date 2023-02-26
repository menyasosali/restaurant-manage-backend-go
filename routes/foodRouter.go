package routes

import (
	"github.com/gorilla/mux"
	controller "github.com/menyasosali/restaurant-manage-backend-go/controllers"
)

func FoodRoutes(incomingRoutes *mux.Router) {
	incomingRoutes.HandleFunc("/foods", controller.GetFoods).Methods("GET")
	incomingRoutes.HandleFunc("/foods/:food_id", controller.GetFood).Methods("GET")
	incomingRoutes.HandleFunc("/foods", controller.CreateFood).Methods("POST")
	incomingRoutes.HandleFunc("/foods/:food_id", controller.UpdateFood).Methods("UPDATE")
}
