package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

func fetchImage(urlStr string) ([]byte, error) {

	url, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("Error parsing url: %v", err)
	}
	// Check remote image size by fetching HTTP Headers
	// MaxAllowedSize := 5 * 1024 * 1000
	// if MaxAllowedSize > 0 {
	// 	var ireq *http.Request
	// 	req := newHTTPRequest(ireq, "HEAD", url)
	// 	res, err := http.DefaultClient.Do(req)
	//
	// 	if err != nil {
	// 		return nil, fmt.Errorf("Error fetching image http headers: %v", err)
	// 	}
	//
	// 	res.Body.Close()
	//
	// 	if res.StatusCode < 200 && res.StatusCode > 206 {
	// 		return nil, fmt.Errorf("Error fetching image http headers: (status=%d) (url=%s)", res.StatusCode, req.URL.String())
	// 	}
	//
	// 	contentLength, _ := strconv.Atoi(res.Header.Get("Content-Length"))
	//
	// 	if contentLength > MaxAllowedSize {
	// 		return nil, fmt.Errorf("Content-Length %d exceeds maximum allowed %d bytes", contentLength, MaxAllowedSize)
	// 	}
	// }

	// Perform the request using the default client
	req, _ := http.NewRequest("GET", url.String(), nil)
	req.Header.Set("User-Agent", "imgserver/1.0.0")
	req.URL = url
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error downloading image: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("Error downloading image: (status=%d) (url=%s)", res.StatusCode, req.URL.String())
	}

	// Read the body
	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("Unable to create image from response body: %s (url=%s)", req.URL.String(), err)
	}
	return buf, nil
}

// func newHTTPRequest(ireq *http.Request, method string, url *url.URL) *http.Request {
// 	req, _ := http.NewRequest(method, url.String(), nil)
// 	req.Header.Set("User-Agent", "imgserver/1.0.0")
// 	req.URL = url
// 	return req
// }
