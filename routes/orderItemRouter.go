package routes

import (
	"github.com/gorilla/mux"
	controller "github.com/menyasosali/restaurant-manage-backend-go/controllers"
)

func OrderItemRoutes(incomingRoutes *mux.Router) {
	incomingRoutes.HandleFunc("/orderItems", controller.GetOrderItems).Methods("GET")
	incomingRoutes.HandleFunc("/orderItems/:orderItem_id", controller.GetOrderItem).Methods("GET")
	incomingRoutes.HandleFunc("/orderItems-order/:order_id", controller.GetOrderItemsByOrder).Methods("GET")
	incomingRoutes.HandleFunc("/orderItems", controller.CreateOrderItem).Methods("POST")
	incomingRoutes.HandleFunc("/orderItems/:orderItem_id", controller.UpdateOrderItem).Methods("UPDATE")
}
