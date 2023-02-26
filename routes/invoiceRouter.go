package routes

import (
	"github.com/gorilla/mux"
	controller "github.com/menyasosali/restaurant-manage-backend-go/controllers"
)

func InvoiceRoutes(incomingRoutes *mux.Router) {
	incomingRoutes.HandleFunc("/invoices", controller.GetInvoices).Methods("GET")
	incomingRoutes.HandleFunc("/invoices/:invoice_id", controller.GetInvoice).Methods("GET")
	incomingRoutes.HandleFunc("/invoices", controller.CreateInvoice).Methods("POST")
	incomingRoutes.HandleFunc("/invoices/:invoice_id", controller.UpdateInvoice).Methods("UPDATE")
}
