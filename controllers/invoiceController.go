package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/menyasosali/restaurant-manage-backend-go/database"
	"github.com/menyasosali/restaurant-manage-backend-go/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"time"
)

type InvoiceViewFormat struct {
	InvoiceId      string
	PaymentMethod  string
	OrderId        string
	PaymentStatus  *string
	PaymentDue     interface{}
	TableNumber    interface{}
	PaymentDueDate time.Time
	OrderDetails   interface{}
}

var invoiceCollection = database.OpenCollection(database.Client, "invoice")

func GetInvoices(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	result, err := invoiceCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		msg := "error occurred while listing invoice items"
		http.Error(w, msg, http.StatusInternalServerError)
	}

	var allInvoices []bson.M
	if err = result.All(ctx, &allInvoices); err != nil {
		log.Fatal(err)
	}

	allInvoicesJSON, err := json.Marshal(allInvoices)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}
	w.WriteHeader(http.StatusOK)
	w.Write(allInvoicesJSON)
}

func GetInvoice(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	vars := mux.Vars(r)
	invoiceId := vars["invoice_id"]

	var invoice models.Invoice
	if err := invoiceCollection.FindOne(ctx, bson.M{"invoice_id": invoiceId}).Decode(&invoice); err != nil {
		msg := "error occurred while listing invoices"
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	var invoiceView InvoiceViewFormat

	allOrderItems, err := ItemsByOrder(invoice.OrderId)
	invoiceView.OrderId = invoice.OrderId
	invoiceView.PaymentDueDate = invoice.PaymentDueDate

	invoiceView.PaymentMethod = "null"
	if invoice.PaymentMethod != nil {
		invoiceView.PaymentMethod = *invoice.PaymentMethod
	}

	invoiceView.InvoiceId = invoice.InvoiceId
	invoiceView.PaymentStatus = *&invoice.PaymentStatus
	invoiceView.PaymentMethod = allOrderItems[0]["payment_due"]
	invoiceView.TableNumber = allOrderItems[0]["table_number"]
	invoiceView.OrderDetails = allOrderItems[0]["order_items"]

	invoiceViewJSON, err := json.Marshal(invoiceView)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(invoiceViewJSON)
}

func CreateInvoice(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var invoice models.Invoice

	if err := json.NewDecoder(r.Body).Decode(&invoice); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var order models.Order

	if err := orderCollection.FindOne(ctx, bson.M{"order_id": invoice.OrderId}).Decode(&order); err != nil {
		msg := fmt.Sprintf("message: Order was not found")
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	status := "PENDING"
	if invoice.PaymentStatus == nil {
		invoice.PaymentStatus = &status
	}

	invoice.PaymentDueDate, _ = time.Parse(time.RFC822, time.Now().AddDate(0, 0, 1).Format(time.RFC822))
	invoice.CreatedAt, _ = time.Parse(time.RFC822, time.Now().Format(time.RFC822))
	invoice.UpdatedAT, _ = time.Parse(time.RFC822, time.Now().Format(time.RFC822))
	invoice.ID = primitive.NewObjectID()
	invoice.OrderId = invoice.ID.Hex()

	validateErr := validate.Struct(invoice)
	if validateErr != nil {
		http.Error(w, validateErr.Error(), http.StatusInternalServerError)
		return
	}

	result, insertErr := invoiceCollection.InsertOne(ctx, invoice)
	if insertErr != nil {
		msg := fmt.Sprintf("Invoice item was not created")
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	resultJson, err := json.Marshal(result)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resultJson)
}

func UpdateInvoice(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var invoice models.Invoice

	vars := mux.Vars(r)
	invoiceId := vars["invoice_id"]

	if err := json.NewDecoder(r.Body).Decode(&invoice); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var updateObj primitive.D

	if invoice.PaymentMethod != nil {
		updateObj = append(updateObj, bson.E{"payment_method", invoice.PaymentMethod})
	}

	if invoice.PaymentStatus != nil {
		updateObj = append(updateObj, bson.E{"payment_status", invoice.PaymentStatus})
	}

	invoice.UpdatedAT, _ = time.Parse(time.RFC822, time.Now().Format(time.RFC822))
	updateObj = append(updateObj, bson.E{"update_at", invoice.UpdatedAT})

	filter := bson.M{"invoice_id": invoiceId}

	upsert := true

	opt := options.UpdateOptions{
		Upsert: &upsert,
	}

	status := "PENDING"
	if invoice.PaymentStatus == nil {
		invoice.PaymentStatus = &status
	}

	result, err := invoiceCollection.UpdateOne(
		ctx,
		filter,
		bson.D{
			{"$set", updateObj},
		},
		&opt,
	)

	if err != nil {
		msg := fmt.Sprintf("invoice item update failed")
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	resultJson, err := json.Marshal(result)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resultJson)
}
