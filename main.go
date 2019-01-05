package main

import (
	"bytes"
	"imgserver/api/app"
	"imgserver/api/controllers"
	"imgserver/imageserver"
	"imgserver/template"
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

	//Account
	router.POST("/api/user/new", controllers.CreateAccount)
	router.POST("/api/user/login", controllers.Authenticate)
	router.POST("/api/user/login/refresh-token", controllers.AuthenticateWithRefreshToken)
	router.GET("/api/user/me", controllers.GetUserProfile)
	router.POST("/api/accounts/forogotpassword", controllers.ForgotPassword)
	router.POST("/api/accounts/resetpassword", controllers.ResetPassword)

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

	// Reports Controller
	router.GET("/api/reports/start/:start/end/:end", controllers.GetReportFor)
	router.GET("/api/reports/logs/start/:start/end/:end/page/:page", controllers.GetLogsFor)

	// router.GET("/", viewHandler)

	router.GET("/", func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		buffer := new(bytes.Buffer)
		template.HomeIndex(buffer)
		w.Write(buffer.Bytes())
	})
	router.GET("/documentation", func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		buffer := new(bytes.Buffer)
		template.GettingStartedIndex(buffer)
		w.Write(buffer.Bytes())
	})
	router.GET("/documentation/upload-image", func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		buffer := new(bytes.Buffer)
		template.UploadObjectIndex(buffer)
		w.Write(buffer.Bytes())
	})
	router.GET("/documentation/modify-image", func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		buffer := new(bytes.Buffer)
		template.ModifyImageIndex(buffer)
		w.Write(buffer.Bytes())
	})

	m := app.JwtAuthentication(router)
	log.Fatal(http.ListenAndServe(":"+port, m))

}
