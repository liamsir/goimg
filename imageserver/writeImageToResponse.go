package imageserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"cloud.google.com/go/storage"
)

var config Config

func init() {

	configFile, err := os.Open("MyProject-89e0f34eb7a6.json")
	defer configFile.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
}

func writeFileToResponseWriter(buf []byte, w http.ResponseWriter) {
	w.Header().Set("Content-Length", strconv.Itoa(len(buf)))
	w.Header().Set("Content-Type", "image/jpeg")
	w.Write(buf)
}
func writeError(w http.ResponseWriter) {
	buf := []byte{}
	w.Header().Set("Content-Length", strconv.Itoa(len(buf)))
	w.Header().Set("Content-Type", "image/jpeg")
	w.Write(buf)
}

func fetchImageAndWriteToResponse(url string, w http.ResponseWriter) {
	contentType := ""
	buf, err := fetchImage(url, &contentType)
	if err != nil {
		return
	}
	writeFileToResponseWriter(buf, w)
}

// func serveImageFromCache(resource map[uint]models.File, w http.ResponseWriter, r *http.Request, usageStats map[int]int, debug bool) (bool, error) {

// 	if cachedResource, ok := resource[1]; ok {
// 		originalResource, ok := resource[0]
// 		if !ok {
// 			return false, fmt.Errorf("Error.")
// 		}
// 		_, err := checkLimit(servedFromCache, usageStats)
// 		if err != nil {
// 			return false, fmt.Errorf("Error.")
// 		}

// 		logRequest(requestEntity{
// 			Body:   resource[0].Name + "/" + resource[1].Name,
// 			FileId: int(resource[1].ID),
// 			UserId: int(resource[1].UserId),
// 			Type:   0,
// 		})
// 		fileName := fmt.Sprintf("%d/%s_/%s",
// 			cachedResource.UserId,
// 			originalResource.Hash,
// 			cachedResource.Hash,
// 		)
// 		url, err := signUrl(fileName)
// 		if err != nil {
// 			return false, err
// 		}
// 		if debug {
// 			fetchImageAndWriteToResponse(url, w)
// 		} else {
// 			http.Redirect(w, r, url, http.StatusTemporaryRedirect)
// 		}

// 		return true, nil
// 	}
// 	return false, nil
// }
func serveImageFromCache(paramUser string, paramResource string, paramModifiers string, w http.ResponseWriter, r *http.Request, debug bool) (bool, error) {

	resourceHash := fmt.Sprint(hash(paramResource))
	version := fmt.Sprint(hash(paramResource + paramModifiers))
	fmt.Println("version", version)
	if value, ok := fileStatus(paramUser, version); ok {

		if !ok {
			return false, fmt.Errorf("Error.")
		}

		if value != 1 {
			return false, nil
		}

		sfc, err := GetUsage(paramUser, servedFromCache)
		if err != nil {
			return false, fmt.Errorf("Error.")
		}
		isfc, _ := strconv.Atoi(sfc)
		_, errc := checkLimit(servedFromCache, map[int]int{servedFromCache: isfc})
		if errc != nil {
			writeError(w)
			return false, fmt.Errorf("Error.")
		}

		fmt.Println(value)
		fmt.Println("resourceHash ", resourceHash)
		fileName := fmt.Sprintf("%s/%s_/%s",
			paramUser,
			resourceHash,
			version,
		)
		url, err := signUrl(fileName)
		fmt.Println("fileName", fileName)
		fmt.Println("url signed", url)
		if err != nil {
			return false, err
		}
		if debug {
			fetchImageAndWriteToResponse(url, w)
		} else {
			http.Redirect(w, r, url, http.StatusTemporaryRedirect)
		}

		return true, nil
	}

	// if cachedResource, ok := resource[1]; ok {
	// 	originalResource, ok := resource[0]
	// 	if !ok {
	// 		return false, fmt.Errorf("Error.")
	// 	}
	// 	_, err := checkLimit(servedFromCache, usageStats)
	// 	if err != nil {
	// 		return false, fmt.Errorf("Error.")
	// 	}

	// 	logRequest(requestEntity{
	// 		Body:   resource[0].Name + "/" + resource[1].Name,
	// 		FileId: int(resource[1].ID),
	// 		UserId: int(resource[1].UserId),
	// 		Type:   0,
	// 	})
	// 	fileName := fmt.Sprintf("%d/%s_/%s",
	// 		cachedResource.UserId,
	// 		originalResource.Hash,
	// 		cachedResource.Hash,
	// 	)
	// 	url, err := signUrl(fileName)
	// 	if err != nil {
	// 		return false, err
	// 	}
	// 	if debug {
	// 		fetchImageAndWriteToResponse(url, w)
	// 	} else {
	// 		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	// 	}

	// 	return true, nil
	// }
	return false, nil
}

type SignUrlParams struct {
	UserId string
	Hash   string
}

func signUrl(fileName string) (string, error) {

	expires := time.Now().Add(time.Second * 15)

	opts := storage.SignedURLOptions{
		GoogleAccessID: config.ClientEmail,
		PrivateKey:     ([]byte(config.PrivateKey)),
		Method:         "GET",
		Expires:        expires,
	}
	url, err := storage.SignedURL(storageBucketName, fileName, &opts)
	if err != nil {
		return "", err
	}
	return url, nil
}

type Config struct {
	ClientEmail string `json:"client_email"`
	PrivateKey  string `json:"private_key"`
}
