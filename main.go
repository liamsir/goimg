package main

import (
	"log"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
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

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"https://goo.gl"},
	})

	handler := c.Handler(router)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
