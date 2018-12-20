package main

import (
	"imgserver/api/app"
	"imgserver/api/controllers"
	"imgserver/imageserver"
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

	router.GET("/user/:user/modifiers/:modifiers/resource/*resource", imageserver.Index)
	router.GET("/user/:user/resource/*resource", imageserver.Index)
	router.GET("/healthz", imageserver.Health)

	router.POST("/api/user/new", controllers.CreateAccount)
	router.POST("/api/user/login", controllers.Authenticate)

	router.POST("/api/file/new", controllers.CreateFile)
	router.GET("/api/user/:id/files", controllers.GetFilesFor)

	router.POST("/user/:user/upload/:signature/file/:fileName", controllers.UploadImage)
	router.POST("/user/:user/signUrl", controllers.SignUrl)

	m := app.JwtAuthentication(router)
	log.Fatal(http.ListenAndServe(":"+port, m))

}
