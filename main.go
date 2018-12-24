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
	router.GET("/api/files/page/:page", controllers.GetFilesFor)
	router.GET("/api/file/:fileId/versions/page/:page", controllers.GetFileVersionsFor)
	router.DELETE("/api/user/:userId/file/:id", controllers.DeleteFile)

	const uploadImageRoute = "/user/:user/upload/:signature/expires/:expires/file/:fileName"
	router.POST(uploadImageRoute, controllers.UploadImage)
	router.PUT(uploadImageRoute, controllers.UploadImage)

	router.POST("/user/:user/signUrl", controllers.SignUrl)

	m := app.JwtAuthentication(router)
	log.Fatal(http.ListenAndServe(":"+port, m))

}
