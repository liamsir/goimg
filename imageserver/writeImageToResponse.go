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
	buf, err := fetchImage(url)
	if err != nil {
		return
	}
	writeFileToResponseWriter(buf, w)
}

func serveImageFromCache(resource map[int32]fileEntity, w http.ResponseWriter, r *http.Request, usageStats map[int]int) (bool, error) {

	if cachedResource, ok := resource[1]; ok {
		originalResource, ok := resource[0]
		if !ok {
			return false, fmt.Errorf("Error.")
		}
		_, err := checkLimit(servedFromCache, usageStats)
		if err != nil {
			return false, fmt.Errorf("Error.")
		}
		cachedUrl := fmt.Sprintf("%s/%d/%s_/%s",
			storageBucketUrl,
			cachedResource.UserId,
			originalResource.Hash,
			cachedResource.Hash,
		)
		fmt.Println("cachedUrl", cachedUrl)

		logRequest(requestEntity{
			Body:   "",
			FileId: resource[1].Id,
			UserId: resource[1].UserId,
			Type:   0,
		})

		// sign url
		fileName := fmt.Sprintf("%d/%s_/%s",
			cachedResource.UserId,
			originalResource.Hash,
			cachedResource.Hash,
		)
		expires := time.Now().Add(time.Second * 15)
		bucketName := "imgmdf"
		//googleAccessID := "image-server@upbeat-aspect-168013.iam.gserviceaccount.com "
		//data, _ := ioutil.ReadFile(serviceAccountPEMFilename)

		opts := storage.SignedURLOptions{
			GoogleAccessID: config.ClientEmail,
			PrivateKey:     ([]byte(config.PrivateKey)),
			Method:         "GET",
			Expires:        expires,
		}
		url, err := storage.SignedURL(bucketName, fileName, &opts)
		if err != nil {
			// TODO: Handle error.
			fmt.Println(err)
		}
		// fmt.Print("\n\n\n")
		// fmt.Println("signed url ", url)
		// fmt.Print("\n\n\n")
		//return true, nil

		// end sign url

		//fetchImageAndWriteToResponse(cachedUrl, w)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
		return true, nil
	}
	return false, nil
}

func serveOriginalImage(resource map[int32]fileEntity, w http.ResponseWriter, usageStats map[int]int) (bool, error) {
	originalResource, ok := resource[0]
	if !ok {
		return false, fmt.Errorf("Error.")
	}
	_, err := checkLimit(servedOriginalImage, usageStats)
	if err != nil {
		return false, fmt.Errorf("Error.")
	}
	originalResourceUrl := fmt.Sprintf("%s/%d/%s",
		storageBucketUrl,
		originalResource.UserId,
		originalResource.Hash,
	)
	fetchImageAndWriteToResponse(originalResourceUrl, w)
	return true, nil
}

type Config struct {
	ClientEmail string `json:"client_email"`
	PrivateKey  string `json:"private_key"`
}
