package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/julienschmidt/httprouter"
)

func handler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	paramModifiers := ps.ByName("modifiers")
	paramResource := ps.ByName("resource")[1:]
	paramUser := ps.ByName("user")

	resource := getResourceInfo(getFileParams{
		userName:  paramUser,
		modifiers: paramModifiers,
		resource:  paramResource,
	})
	fmt.Println("getResourceInfo0 ", resource[0].Id)
	fmt.Println("getResourceInfo1 ", resource[1].Id)

	// 1. Serve image from cache
	servedFromCache, e := serveImageFromCache(resource, w)

	if e != nil || servedFromCache {
		fmt.Println("Served from cache.")
		return
	}

	if paramModifiers == "" {
		servedOriginalImage, e := serveOriginalImage(resource, w)
		if e != nil || servedOriginalImage {
			fmt.Println("Served original image.")
			return
		}
	}

	fmt.Println("Beginning to perform image transformation...")
	var originalResourceUrl string

	if originalResource, ok := resource[0]; ok {
		// 2. Check if user has permission to perform operations
		if operationsAllowed(originalResource.AllowedOperations, paramModifiers) {
			originalResourceUrl =
				fmt.Sprintf("%s/%d/%s",
					storageBucketUrl,
					originalResource.UserId,
					originalResource.Hash,
				)
		} else {
			fmt.Println("Operation is not allowed.")
			return
		}
		fmt.Println("originalResource ", originalResource.Hash)
	} else {
		fmt.Println("Resource doesn't exists.")
		/*
			4. If resource doesn't exists, check if resource is url
			If it's url, then try to download the resource and save in
			blob storage and in db
		*/
		_, err := url.ParseRequestURI(paramResource)
		if err != nil {
			return
		}
		fmt.Println("Downloading remote resource and storing...")
		originalResourceUrl, err = downloadResourceAndSaveInBlob(
			downloadAndSaveObjectParams{
				ResourceUrl: paramResource,
				UserName:    paramUser,
			},
		)
		fmt.Println("remote resource ", originalResourceUrl)
		if err != nil {
			fmt.Println("failed to fetch remote resource", originalResourceUrl)
			return
		}
	}

	fmt.Println("performing operations...", originalResourceUrl)
	_, err := performOperationsAndWriteImageToRequest(
		performOperationsParam{
			resourceUrl:    originalResourceUrl,
			paramModifiers: paramModifiers,
			userName:       paramUser,
			resource:       paramResource,
		}, w)
	fmt.Println("end of performing operations")
	if err != nil {
		fmt.Println("failed to perform operations")
		return
	}
}

func healthController(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	health := GetHealthStats()
	body, _ := json.Marshal(health)
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}
func main() {

	port := os.Getenv("PORT")

	if port == "" {
		port = "3001"
		// log.Fatal("$PORT must be set")
	}

	router := httprouter.New()

	router.GET("/user/:user/modifiers/:modifiers/resource/*resource", handler)
	router.GET("/user/:user/resource/*resource", handler)
	router.GET("/health", healthController)

	log.Fatal(http.ListenAndServe(":"+port, router))
}
