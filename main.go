package main

import (
	router2 "go-server-server-generated/go/router"
	"log"
	"net/http"
)

func main() {
	log.Printf("Server started")

	router := router2.NewRouter()
	log.Println("started 5000")
	http.ListenAndServe(":5000", router)
}
