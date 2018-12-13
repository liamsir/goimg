package main

import (
	"log"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
)

func main() {

	port := os.Getenv("PORT")

	if port == "" {
		port = "3001"
		// log.Fatal("$PORT must be set")
	}

	router := httprouter.New()

	router.GET("/user/:user/modifiers/:modifiers/resource/*resource", index)
	router.GET("/user/:user/resource/*resource", index)
	router.GET("/healthz", health)

	log.Fatal(http.ListenAndServe(":"+port, router))
}
