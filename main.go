package main

import (
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/menyasosali/restaurant-manage-backend-go/middleware"
	"github.com/menyasosali/restaurant-manage-backend-go/routes"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8000"
	}

	router := mux.NewRouter()

	router.Use(func(h http.Handler) http.Handler {
		return handlers.LoggingHandler(os.Stdout, h)
	})
	routes.UserRoutes(router)
	router.Use(middleware.Authentication)

	routes.FoodRoutes(router)
	routes.MenuRoutes(router)
	routes.TableRoutes(router)
	routes.OrderRoutes(router)
	routes.OrderItemRoutes(router)
	routes.InvoiceRoutes(router)

	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Panicf("cannot start server on port %s: %s", port, err)
	}
}
