package routes

import (
	"github.com/gorilla/mux"
	controller "github.com/menyasosali/restaurant-manage-backend-go/controllers"
)

func UserRoutes(incomingRoutes *mux.Router) {
	incomingRoutes.HandleFunc("/users", controller.GetUsers).Methods("GET")
	incomingRoutes.HandleFunc("/users/:user_id", controller.GetUser).Methods("GET")
	incomingRoutes.HandleFunc("/users/signup", controller.SingUp).Methods("POST")
	incomingRoutes.HandleFunc("/users/login", controller.Login).Methods("POST")
}
