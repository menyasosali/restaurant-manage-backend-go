package routes

import (
	"github.com/gorilla/mux"
	controller "github.com/menyasosali/restaurant-manage-backend-go/controllers"
)

func TableRoutes(incomingRoutes *mux.Router) {
	incomingRoutes.HandleFunc("/tables", controller.GetTables).Methods("GET")
	incomingRoutes.HandleFunc("/tables/:table_id", controller.GetTable).Methods("GET")
	incomingRoutes.HandleFunc("/tables", controller.CreateTable).Methods("POST")
	incomingRoutes.HandleFunc("/tables/:table_id", controller.UpdateTable).Methods("UPDATE")
}
