package routes

import (
	"github.com/gorilla/mux"
	controller "github.com/menyasosali/restaurant-manage-backend-go/controllers"
)

func MenuRoutes(incomingRoutes *mux.Router) {
	incomingRoutes.HandleFunc("/menus", controller.GetMenus).Methods("GET")
	incomingRoutes.HandleFunc("/menus/:menu_id", controller.GetMenu).Methods("GET")
	incomingRoutes.HandleFunc("/menus", controller.CreateMenu).Methods("POST")
	incomingRoutes.HandleFunc("/menus/:menu_id", controller.UpdateMenu).Methods("UPDATE")
}
