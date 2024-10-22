package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

var tracer = otel.Tracer("Groceries API")

var groceries = []Grocery{
	{Name: "Almod Milk", Quantity: 2},
	{Name: "Apple", Quantity: 6},
}

func AllGroceries(w http.ResponseWriter, r *http.Request) {
	_, span := tracer.Start(r.Context(), "get all groceries")
	defer span.End()

	method := r.Method
	scheme := "http"
	statusCode := 200
	host := r.Host
	port := r.URL.Port()
	if port == "" {
		port = "8081"
	}

	// Set span status
	span.SetStatus(codes.Ok, "")

	// Use semantic conventions for common attributes
	span.SetAttributes(
		semconv.HTTPMethodKey.String(method),
		semconv.HTTPSchemeKey.String(scheme),
		semconv.HTTPStatusCodeKey.Int(statusCode),
		semconv.HTTPTargetKey.String(r.URL.Path),
		semconv.HTTPURLKey.String(r.URL.String()),
		semconv.HTTPHostKey.String(host),
		semconv.NetHostPortKey.String(port),
		semconv.HTTPUserAgentKey.String(r.UserAgent()),
		semconv.HTTPRequestContentLengthKey.Int64(r.ContentLength),
		semconv.NetPeerIPKey.String(r.RemoteAddr),
	)

	// Custom attributes that don't have semantic conventions
	startTime := time.Now()
	span.SetAttributes(
		attribute.String("created_at", startTime.Format(time.RFC3339Nano)),
		attribute.Float64("duration_ns", float64(time.Since(startTime).Nanoseconds())),
		attribute.String("parent_id", ""), // You might need to extract this from the context
		attribute.String("referer", r.Referer()),
		attribute.String("request_type", "Incoming"),
		attribute.String("sdk_type", "go-mux"),
		attribute.String("service_version", ""), // Fill in your service version if available
		attribute.StringSlice("tags", []string{}),
	)

	// Set nested fields (these don't have direct semconv equivalents)
	span.SetAttributes(
		attribute.String("query_params", fmt.Sprintf("%v", r.URL.Query())),
		attribute.String("request_body", "{}"), // Assuming empty body for GET request
		attribute.String("request_headers", fmt.Sprintf("%v", r.Header)),
		attribute.String("response_body", "{}"),
		attribute.String("response_headers", "{}"),
	)

	fmt.Println("Endpoint hit: returnAllGroceries")
	json.NewEncoder(w).Encode(groceries)
	if span.IsRecording() {
		span.SetAttributes(
			attribute.Int64("enduser.id", 1),
			attribute.String("enduser.email", "user.Email"),
			attribute.String("http.method", "GET"),
			attribute.String("http.route", "/allgroceries"),
		)
	}
}

func SingleGrocery(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	name := vars["name"]

	for _, grocery := range groceries {
		if grocery.Name == name {
			json.NewEncoder(w).Encode(grocery)
		}
	}
}

func GroceriesToBuy(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)
	var grocery Grocery
	json.Unmarshal(reqBody, &grocery)
	groceries = append(groceries, grocery)

	json.NewEncoder(w).Encode(groceries)

}

func DeleteGrocery(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	name := vars["name"]

	for index, grocery := range groceries {
		if grocery.Name == name {
			groceries = append(groceries[:index], groceries[index+1:]...)
		}
	}

}
func UpdateGrocery(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	name := vars["name"]

	for index, grocery := range groceries {
		if grocery.Name == name {
			groceries = append(groceries[:index], groceries[index+1:]...)

			var updateGrocery Grocery

			json.NewDecoder(r.Body).Decode(&updateGrocery)
			groceries = append(groceries, updateGrocery)
			fmt.Println("Endpoint hit: UpdateGroceries")
			json.NewEncoder(w).Encode(updateGrocery)
			return
		}
	}

}
