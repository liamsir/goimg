package imageserver

import (
	"fmt"
	"net/http"
	"strconv"
)

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

func serveImageFromCache(resource map[int32]fileEntity, w http.ResponseWriter, usageStats map[int]int) (bool, error) {

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
		fetchImageAndWriteToResponse(cachedUrl, w)
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
