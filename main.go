package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"hash/fnv"
	"image/jpeg"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/lib/pq"
	"github.com/nfnt/resize"
)

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func userHasPermissionsToPerformOperations(
	allowedOperations string,
	inputOperations string) bool {
	return true
}

func handler(w http.ResponseWriter, r *http.Request) {
	routePath := r.URL.String()
	routeSplited := strings.Split(routePath, "modifiers")

	var paramModifiers = ""
	var paramResource = ""
	var paramUser = ""

	if len(routeSplited) > 1 {
		paramModifiers = routeSplited[1][1:]
		resourceSplited := strings.Split(routeSplited[0], "resource")
		if len(resourceSplited) > 1 {
			paramResource = resourceSplited[1][1 : len(resourceSplited[1])-1]
			userSplited := strings.Split(resourceSplited[0], "user")
			if len(userSplited) > 1 {
				paramUser = userSplited[1][1 : len(userSplited[1])-1]
			}
		}
	} else {
		return
	}

	user := paramUser
	resourceModifiers := paramResource + paramModifiers
	resModHash := fmt.Sprint(hash(resourceModifiers))
	resHash := fmt.Sprint(hash(paramResource))

	// 1. Check if resource exists in cache for given parameters
	connStr := "postgres://jxbnzxtecqvcsv:9f603a3b7a60b5583f668fa2cf0ab0badd2c8f9dbacc73564cb1e9ee45241312@ec2-54-246-85-234.eu-west-1.compute.amazonaws.com:5432/dag2mo4a48vlb3"
	db, err := sql.Open("postgres", connStr)
	rows, err := db.Query(`select f.hash, uf.user_id, f.type, f.allowed_operations from file f
join user_file uf on uf.file_id = f.id
where f.hash = $1 or f.hash = $2	and uf.user_id = (select id from "user" where username = $3)`,
		resHash, resModHash, user)
	if err, ok := err.(*pq.Error); ok {
		fmt.Println("pq error:", err.Code.Name())
	}
	fmt.Println("resHash", resHash)
	fmt.Println("resModHash", resModHash)
	var orHash string
	var mdfHash string
	var userId int32
	var ftype int32
	var allowedOperations string

	for rows.Next() {
		var hash string
		var ao sql.NullString
		err = rows.Scan(&hash, &userId, &ftype, &ao)
		if err != nil {
			log.Fatal(err)
		}

		if ftype == 0 {
			orHash = hash
			if ao.Valid {
				allowedOperations = ao.String
			}
		}

		if ftype == 1 {
			mdfHash = hash
		}
	}

	if mdfHash != "" {
		// get file from storage using hash and return as image
		providerUrl := fmt.Sprintf(
			"https://storage.googleapis.com/imgmdf/%d/%s",
			userId,
			mdfHash)

		u, err := url.Parse(providerUrl)
		if err != nil {
			panic(err)
		}
		buf, err := fetchImage(u, r)

		if err != nil {
			return
		}
		w.Header().Set("Content-Length", strconv.Itoa(len(buf)))
		w.Header().Set("Content-Type", "image/jpeg")
		w.Write(buf)
	}

	// original resource
	// 2. Check if user has permissions to perform transformations
	// add allowed operations in file table on hash form
	//...
	fmt.Print("orHash ", orHash)
	fmt.Println("allowedOperations", allowedOperations)

	if userHasPermissionsToPerformOperations(allowedOperations, paramModifiers) {
		// 3. Check if resource exists for user
		var providerUrl string
		if orHash != "" {
			fmt.Println("original resource hash %s", orHash)
			providerUrl = fmt.Sprintf(
				"https://storage.googleapis.com/imgmdf/%d/%s",
				userId,
				orHash)
		} else {
			// 4. If resource doesn't exists, check if resource is url
			// If it's url, then try to download the resource and save in
			// blob storage and in db
			providerUrl = paramResource
		}

		u, err := url.Parse(providerUrl)
		if err != nil {
			panic(err)
		}
		fmt.Print(providerUrl)

		buf, err := fetchImage(u, r)
		if err != nil {
			panic(err)
		}
		//fmt.Print(buf)
		r := bytes.NewReader(buf)
		img, err := jpeg.Decode(r)
		if err != nil {
			log.Fatal(err)
		}

		modifiers := strings.Split(resourceModifiers, "_")
		if len(modifiers) > 0 {
			widthStr := modifiers[1]
			heightStr := modifiers[2]
			width, e := strconv.ParseUint(strings.Replace(widthStr, "w", "", -1), 10, 32)
			height, e := strconv.ParseUint(strings.Replace(heightStr, "h", "", -1), 10, 32)

			if e != nil {
				log.Fatal(err)
			}

			// 5. Perform transformations, and save transformed image to blob and db
			newImage := resize.Resize(uint(width), uint(height), img, resize.Lanczos3)

			bufOut := new(bytes.Buffer)
			err = jpeg.Encode(bufOut, newImage, nil)
			sendBuf := bufOut.Bytes()
			// 6. Return edited image
			w.Header().Set("Content-Length", strconv.Itoa(len(sendBuf)))
			w.Header().Set("Content-Type", "image/jpeg")
			w.Write(sendBuf)
		}

	}
}

func fetchImage(url *url.URL, ireq *http.Request) ([]byte, error) {
	// Check remote image size by fetching HTTP Headers
	MaxAllowedSize := 5 * 1024 * 1000
	if MaxAllowedSize > 0 {
		req := newHTTPRequest(ireq, "HEAD", url)
		res, err := http.DefaultClient.Do(req)

		if err != nil {
			return nil, fmt.Errorf("Error fetching image http headers: %v", err)
		}

		res.Body.Close()

		if res.StatusCode < 200 && res.StatusCode > 206 {
			return nil, fmt.Errorf("Error fetching image http headers: (status=%d) (url=%s)", res.StatusCode, req.URL.String())
		}

		contentLength, _ := strconv.Atoi(res.Header.Get("Content-Length"))

		if contentLength > MaxAllowedSize {
			return nil, fmt.Errorf("Content-Length %d exceeds maximum allowed %d bytes", contentLength, MaxAllowedSize)
		}
	}

	// Perform the request using the default client
	req := newHTTPRequest(ireq, "GET", url)
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

func newHTTPRequest(ireq *http.Request, method string, url *url.URL) *http.Request {
	req, _ := http.NewRequest(method, url.String(), nil)
	req.Header.Set("User-Agent", "imgserver/1.0.0")
	req.URL = url
	return req
}

func main() {
	port := os.Getenv("PORT")
	log.Println(port)
	if port == "" {
		port = "3001"
		log.Fatal("$PORT must be set")
	}
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":"+port, nil))
	// router := httprouter.New()
	// router.GET("/user/:user/resource/:resource/modifiers/:modifiers", handler)
	// router.GET("/user/:user/resource/:resource/", handler)
	// log.Fatal(http.ListenAndServe(":"+port, router))
}
