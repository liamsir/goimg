package imageserver

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	observable "github.com/GianlucaGuarini/go-observable"
	"github.com/julienschmidt/httprouter"
)

func extractResourceFromRequestURI(r string) string {
	uriSplited := strings.Split(r, "resource")
	if len(uriSplited) > 0 {
		uri := uriSplited[len(uriSplited)-1][1:]
		return uri
	}
	return ""
}

var files = map[string]int{}
var o *observable.Observable = observable.New()

func Index(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	notifier, ok := w.(http.CloseNotifier)
	if !ok {
		panic("expected http.ResponseWriter to be an http.CloseNotifier")
	}
	// debugMode := false
	//
	// if r.Referer() == "" {
	// 	debugMode = true
	// }
	// paramUser := ps.ByName("user")
	paramModifiers := ps.ByName("modifiers")
	paramResource := extractResourceFromRequestURI(r.RequestURI)
	version := fmt.Sprint(hash(paramResource + paramModifiers))

	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan string)

	onReady := func() {
		fmt.Println("done" + version)
		close(ch)
		ctx.Done()
	}

	if value, ok := files[version]; ok {
		if value == 1 {
			go func() {
				o.On("done"+version, onReady)
			}()
		}
		if value == 2 {
			fmt.Println("value 2")
			fmt.Fprint(w, "processing completed")
			close(ch)
			ctx.Done()
		}
	} else {
		files[version] = 1
		go func() {
			select {
			case <-time.After(time.Second * 5):
				files[version] = 2
				ch <- "Successful result."
				fmt.Println("value 3 done", version)
				o.Trigger("done" + version)
				ctx.Done()
			case <-ctx.Done():
				close(ch)
			}
		}()
	}
	select {
	case result := <-ch:
		fmt.Fprint(w, result)
		cancel()
		ctx.Done()
		return
	case <-notifier.CloseNotify():
		o.Off("done"+version, onReady)
		fmt.Println("Client has disconnected.")
	}
	cancel()
	<-ch
}

// var files = map[int]int{}
// var o *observable.Observable = observable.New()
//
// func Index(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
//
// 	ctx, cancel := context.WithCancel(context.Background())
// 	ch := make(chan string)
//
// 	if value, ok := files[1]; ok {
// 		if value == 1 {
// 			go func() {
// 				o.On("done", func() {
// 					fmt.Println("value 1 processed")
// 					close(ch)
// 					ctx.Done()
// 				})
// 			}()
// 		}
// 		if value == 2 {
// 			fmt.Println("value 2")
// 			fmt.Fprint(w, "processing completed")
// 			close(ch)
// 			ctx.Done()
// 		}
// 	} else {
// 		files[1] = 1
// 		go func() {
// 			select {
// 			case <-time.After(time.Second * 3):
// 				files[1] = 2
// 				ch <- "Successful result."
// 				fmt.Println("value 3 done")
// 				o.Trigger("done")
// 				ctx.Done()
// 			case <-ctx.Done():
// 				close(ch)
// 			}
// 		}()
//
// 	}
//
// 	// go func() {
// 	// 	select {
// 	// 	case <-time.After(time.Second * 3):
// 	// 		ch <- "Successful result."
// 	// 	case <-ctx.Done():
// 	// 		close(ch)
// 	// 	}
// 	// }()
//
// 	select {
// 	case result := <-ch:
// 		fmt.Fprint(w, result)
// 		cancel()
// 		ctx.Done()
// 		return
// 	}
//
// }

// var files = map[int]int{}
// var o *observable.Observable = observable.New()
//
// func Index(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
//
// 	//ctx := r.Context()
// 	log.Printf("handler started")
// 	defer log.Printf("hander ended")
// 	// select {
// 	// case <-time.After(5 * time.Second):
// 	// 	fmt.Println(w, "hello")
// 	// case <-ctx.Done():
// 	// 	log.Print(ctx.Err())
// 	// 	http.Error(w, ctx.Err().Error(), http.StatusInternalServerError)
// 	// }
// 	ctx, cancel := context.WithCancel(context.Background())
//
// 	if value, ok := files[1]; ok {
// 		if value == 1 {
// 			o.On("ready", func(message string, ctx context.Context, ch chan<- string) {
// 				fmt.Println("ready", message)
// 				fmt.Fprint(w, "sfsdfsdf")
// 				cancel()
// 				fmt.Println(ctx)
// 				ctx.Done()
// 				return
// 			})
// 			fmt.Println("processing")
// 		}
// 		if value == 2 {
// 			fmt.Println("processing completed")
// 			fmt.Fprint(w, "processing completed")
// 			cancel()
// 		}
// 	}
//
// 	ch := make(chan string)
//
// 	// doesn't exists in cache, perform operation
// 	if _, ok := files[1]; !ok {
// 		o.On("ready", func(message string, ctx context.Context, ch chan<- string) {
// 			fmt.Println("ready", message)
// 			fmt.Fprint(w, message)
// 			cancel()
// 			return
// 		})
// 		files[1] = 1 // is processing
// 		go longOperation(ctx, ch, o)
// 		// select {
// 		// case <-time.After(10 * time.Second):
// 		// 	o.Trigger("ready", "done")
// 		// 	files[1] = 2 // done
// 		// 	fmt.Println("timeout 1")
// 		// }
// 	}
// 	select {
// 	case result := <-ch:
// 		o.Trigger("ready", result, ctx, ch)
// 		fmt.Fprint(w, result)
// 		cancel()
// 		return
// 	}
// 	cancel()
// 	<-ch
//
// }
//
// func longOperation(ctx context.Context, ch chan<- string, o *observable.Observable) {
// 	// Simulate long operation.
// 	// Change it to more than 10 seconds to get server timeout.
// 	select {
// 	case <-time.After(time.Second * 3):
// 		// o.Trigger("ready", "done", ctx, ch)
// 		files[1] = 2
// 		ch <- "Successful result."
// 	case <-ctx.Done():
// 		close(ch)
// 	}
// }

func Indexsdf(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	debugMode := false

	if r.Referer() == "" {
		debugMode = true
	}
	paramUser := ps.ByName("user")
	paramModifiers := ps.ByName("modifiers")
	paramResource := extractResourceFromRequestURI(r.RequestURI) //ps.ByName("resource")[1:]

	//validate origin
	errOrigin := CheckOrigin(CheckOriginParams{
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

	if len(resource) == 0 {
		writeError(w)
		return
	}

	usageStats := getUsage(paramUser)

	// 1. Serve image from cache
	servedFromCache, e := serveImageFromCache(resource, w, r, usageStats, debugMode)

	if e != nil {
		return
	}

	if servedFromCache {
		fmt.Println("Served from cache.")
		return
	}

	if paramModifiers == "" {
		servedOriginalImage, e := serveOriginalImage(resource, w, r, usageStats, debugMode)
		if e != nil || servedOriginalImage {
			logRequest(requestEntity{
				Body:   resource[0].Name,
				FileId: int(resource[0].ID),
				UserId: int(resource[0].UserId),
				Type:   1,
			})
			fmt.Println("Served original image.")
			return
		}
	}

	fmt.Println("Beginning to perform image transformation...")
	modifiers, err := parseModifiers(paramModifiers)
	if err != nil {
		writeError(w)
		return
	}
	var originalResourceUrl string
	var userId int
	var fileId int
	var fileHash string

	if originalResource, ok := resource[0]; ok {
		// 2. Check if user has permission to perform operations
		if operationsAllowed(originalResource.AllowedOperations, paramModifiers) {
			originalResourceUrl =
				fmt.Sprintf("%s/%d/%s",
					storageBucketUrl,
					originalResource.UserId,
					originalResource.Hash,
				)
			userId = int(originalResource.UserId)
			fileId = int(originalResource.ID)
			fileHash = originalResource.Hash
		} else {
			fmt.Println("Operation is not allowed.")
			return
		}
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
			fmt.Println(errRemoteOrigin)
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
		fileHash = newFile.Hash
		originalResourceUrl = newFile.ResourceURL
		fmt.Println("remote resource ", originalResourceUrl)
		if err != nil {
			fmt.Println("failed to fetch remote resource", originalResourceUrl)
			writeError(w)
			return
		}
	}

	fmt.Println("performing operations...", originalResourceUrl)
	status, err := performOperationsAndWriteImageToRequest(
		performOperationsParam{
			resourceUrl:     originalResourceUrl,
			modifiers:       modifiers,
			modifiersString: paramModifiers,
			userName:        paramUser,
			resource:        paramResource,
			userId:          userId,
			resourceHash:    fileHash,
			fileId:          fileId,
		}, w, usageStats)
	if !status || err != nil {
		writeError(w)
		return
	}
	logRequest(requestEntity{
		Body:   resource[0].Name + "/" + paramModifiers,
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
