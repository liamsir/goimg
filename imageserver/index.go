package imageserver

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

func Index(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	paramUser := ps.ByName("user")
	paramModifiers := ps.ByName("modifiers")
	paramResource := extractResourceFromRequestURI(r.RequestURI) //ps.ByName("resource")[1:]

	//validate origin
	errOrigin := checkOrigin(checkOriginParams{
		UserName: paramUser,
		Request:  r,
	})

	if errOrigin != nil {
		writeError(w)
		return
	}

	resource := getResourceInfo(getFileParams{
		userName:  paramUser,
		modifiers: paramModifiers,
		resource:  paramResource,
	})

	usageStats := getUsage(paramUser)

	fmt.Println("getResourceInfo0 ", resource[0].Id)
	fmt.Println("getResourceInfo1 ", resource[1].Id)

	// 1. Serve image from cache
	servedFromCache, e := serveImageFromCache(resource, w, r, usageStats)

	if e != nil {
		return
	}

	if servedFromCache {
		fmt.Println("Served from cache.")
		return
	}

	if paramModifiers == "" {
		servedOriginalImage, e := serveOriginalImage(resource, w, usageStats)
		if e != nil || servedOriginalImage {
			logRequest(requestEntity{
				Body:   "",
				FileId: resource[0].Id,
				UserId: resource[0].UserId,
				Type:   1,
			})
			fmt.Println("Served original image.")
			return
		}
	}

	fmt.Println("Beginning to perform image transformation...")
	var originalResourceUrl string
	var userId int
	var fileId int
	if originalResource, ok := resource[0]; ok {
		// 2. Check if user has permission to perform operations
		if operationsAllowed(originalResource.AllowedOperations, paramModifiers) {
			originalResourceUrl =
				fmt.Sprintf("%s/%d/%s",
					storageBucketUrl,
					originalResource.UserId,
					originalResource.Hash,
				)
			userId = originalResource.UserId
			fileId = originalResource.Id
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
		errRemoteOrigin := checkRemoteOrigin(checkRemoteOriginParams{
			UserName: paramUser,
			UrlStr:   paramResource,
		})

		if errRemoteOrigin != nil {
			writeError(w)
			return
		}

		_, err := url.ParseRequestURI(paramResource)
		if err != nil {
			return
		}

		fmt.Println("Downloading remote resource and storing...")
		newFile, err := downloadResourceAndSaveInBlob(
			downloadAndSaveObjectParams{
				ResourceUrl: paramResource,
				UserName:    paramUser,
			},
			usageStats,
		)
		userId = newFile.UserId
		fileId = newFile.Id
		originalResourceUrl = newFile.ResourceURL
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
		}, w, usageStats)
	logRequest(requestEntity{
		Body:   paramModifiers,
		FileId: fileId,
		UserId: userId,
		Type:   3,
	})
	fmt.Println("end of performing operations")
	if err != nil {
		fmt.Println("failed to perform operations")
		return
	}
}
