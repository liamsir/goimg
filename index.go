package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/julienschmidt/httprouter"
)

func extractResourceFromRequestURI(r string) string {
	uriSplited := strings.Split(r, "resource")
	if len(uriSplited) > 0 {
		return uriSplited[len(uriSplited)-1][1:]
	}
	return ""
}
func index(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	paramUser := ps.ByName("user")
	paramModifiers := ps.ByName("modifiers")
	paramResource := extractResourceFromRequestURI(r.RequestURI) //ps.ByName("resource")[1:]

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
			writeError(w)
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
