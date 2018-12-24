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

	router.POST("/api/files/new", controllers.CreateFile)
	router.GET("/api/files/page/:page", controllers.GetFilesFor)
	router.GET("/api/file/:id/versions/page/:page", controllers.GetFileVersionsFor)
	router.DELETE("/api/files/:id", controllers.DeleteFile)

	const uploadImageRoute = "/user/:user/upload/:signature/expires/:expires/file/:fileName"
	router.POST(uploadImageRoute, controllers.UploadImage)
	router.PUT(uploadImageRoute, controllers.UploadImage)

	router.POST("/user/:user/signUrl", controllers.SignUrl)

	// Allowed Domains
	router.GET("/api/domains", controllers.GetDomainsFor)
	router.GET("/api/domains/:id", controllers.GetDomain)
	router.POST("/api/domains", controllers.CreateDomain)
	router.PUT("/api/domains/:id", controllers.UpdateDomain)
	router.PATCH("/api/domains/:id", controllers.PatchDomain)
	router.DELETE("/api/domains/:id", controllers.DeleteDomain)

	m := app.JwtAuthentication(router)
	log.Fatal(http.ListenAndServe(":"+port, m))

}
