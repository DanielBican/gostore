package main

import (
	"log"
	"net/http"

	"github.com/DanielBican/gostore/database"
	"github.com/DanielBican/gostore/handlers"
	"github.com/gorilla/mux"
)

const (
	addr = ":8080"
)

func main() {

	// Create a DB connection
	db, err := database.Open()
	if err != nil {
		log.Panic(err)
	}

	// Seed the DB
	database.Seed(db, database.Users, database.Products)

	// Initialize handlers
	h := handlers.HandleGroup{db}
	r := mux.NewRouter()
	r.HandleFunc("/v1/login", h.Login)
	r.HandleFunc("/v1/logout", h.Logout)
	r.HandleFunc("/v1/cart", h.AddToCart)
	r.HandleFunc("/v1/checkout", h.Checkout)

	log.Printf("starting to listen for HTTP calls on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
