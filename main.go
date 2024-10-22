package main

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func main() {

	_db, err := gorm.Open(sqlite.Open("./test.db"), &gorm.Config{})

	if err != nil {
		panic(err)
	}
	db = _db

	if err := db.AutoMigrate(&Grocery{}); err != nil {
		panic(err)
	}

	r := mux.NewRouter().StrictSlash(true)

	ctx := context.Background()

	// Set up OpenTelemetry.
	otelShutdown, err := setupOTelSDK(ctx)
	if err != nil {
		return
	}
	// Handle shutdown properly so nothing leaks.
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	r.HandleFunc("/allgroceries", AllGroceries)
	r.HandleFunc("/groceries/{name}", SingleGrocery)
	r.HandleFunc("/groceries", GroceriesToBuy).Methods("POST")
	r.HandleFunc("/groceries/{name}", UpdateGrocery).Methods("PUT")
	r.HandleFunc("/groceries/{name}", DeleteGrocery).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":10000", r))
}
