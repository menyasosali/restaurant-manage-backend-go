package routes

import (
	"github.com/gorilla/mux"
	controller "github.com/menyasosali/restaurant-manage-backend-go/controllers"
)

func OrderRoutes(incomingRoutes *mux.Router) {
	incomingRoutes.HandleFunc("/orders", controller.GetOrders).Methods("GET")
	incomingRoutes.HandleFunc("/orders/:order_id", controller.GetOrder).Methods("GET")
	incomingRoutes.HandleFunc("/orders", controller.CreateOrder).Methods("POST")
	incomingRoutes.HandleFunc("/orders/:order_id", controller.UpdateOrder).Methods("UPDATE")
}
