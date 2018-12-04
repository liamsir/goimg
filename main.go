package main

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"hash/fnv"
	"image/jpeg"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/lib/pq"
	"github.com/nfnt/resize"
	"google.golang.org/api/option"
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

func handler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	paramModifiers := ps.ByName("modifiers")
	paramResource := ps.ByName("resource")[1:]
	paramUser := ps.ByName("user")

	fmt.Println("querying db...")
	user := paramUser
	resourceModifiers := paramResource + paramModifiers
	resModHash := fmt.Sprint(hash(resourceModifiers))
	resHash := fmt.Sprint(hash(paramResource))
	// 1. Check if resource exists in cache for given parameters
	connStr := "postgres://jxbnzxtecqvcsv:9f603a3b7a60b5583f668fa2cf0ab0badd2c8f9dbacc73564cb1e9ee45241312@ec2-54-246-85-234.eu-west-1.compute.amazonaws.com:5432/dag2mo4a48vlb3"
	db, err := sql.Open("postgres", connStr)
	rows, err := db.Query(`select f.hash, uf.user_id, f.type, f.allowed_operations from file f
join user_file uf on uf.file_id = f.id
where (f.hash = $1 and f.type = 0) or f.hash = $2 and uf.user_id = (select id from "user" where username = $3)`,
		resHash, resModHash, user)
	if err, ok := err.(*pq.Error); ok {
		fmt.Println("pq error:", err.Code.Name())
	}

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
	fmt.Println("end querying db")
	if mdfHash != "" || paramModifiers == "" {
		fmt.Println("downloading resource...")
		// get file from storage using hash and return as image
		providerUrl := fmt.Sprintf(
			"https://storage.googleapis.com/imgmdf/%d/%s",
			userId,
			orHash,
		)
		if mdfHash != "" {
			providerUrl = fmt.Sprintf("%s_/%s", providerUrl, mdfHash)
		}
		u, err := url.Parse(providerUrl)
		if err != nil {
			panic(err)
		}

		buf, err := fetchImage(u)

		if err != nil {
			return
		}
		w.Header().Set("Content-Length", strconv.Itoa(len(buf)))
		w.Header().Set("Content-Type", "image/jpeg")
		w.Write(buf)
		db.Close()
		fmt.Println("served from cache")
		fmt.Println("done")
		return
	}

	// original resource
	// 2. Check if user has permissions to perform transformations
	// add allowed operations in file table on hash form
	//...

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
			fmt.Println("resource does not exist", paramResource)
			u, err := url.Parse(paramResource)
			if err != nil {
				panic(err)
			}
			buf, err := fetchImage(u)
			if err != nil {
				panic(err)
			}
			fmt.Print(len(buf))
			r := bytes.NewReader(buf)
			ctx := context.Background()
			client, err := storage.NewClient(ctx, option.WithCredentialsFile("MyProject-89e0f34eb7a6.json"))

			if err != nil {
				log.Fatalf("Failed to create client: %v", err)
			}
			// Sets the name for the new bucket.
			bucketName := "imgmdf"
			// Creates a Bucket instance.
			bucket := client.Bucket(bucketName)
			obj := bucket.Object("1/" + resHash)
			wc := obj.NewWriter(ctx)

			if _, err = io.Copy(wc, r); err != nil {
				panic(err)
			}
			if err := wc.Close(); err != nil {
				panic(err)
			}

			acl := obj.ACL()
			if err := acl.Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
				panic(err)
			}
			sqlStatement := `
			INSERT INTO public.file
				(created_at, modified_at, last_opened, guid, hash, "index", parent_id, visible, "name", description, "type", allowed_operations, performed_operations)
				VALUES(now(), now(), now(), $1, $2, 0, null, true, $3, '', 0, '', $4) RETURNING id;`
			fileGuid, err := uuid.NewRandom()
			fmt.Print("fileGuid", fileGuid)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("%s", fileGuid)
			var newFileId int
			errIn := db.QueryRow(sqlStatement, fileGuid, resHash, paramResource, resourceModifiers).Scan(&newFileId)
			if errIn != nil {
				log.Fatal(errIn)
			}
			fmt.Printf("%s", fileGuid)

			insertUserFile := `
			INSERT INTO public.user_file
			(user_id,
				file_id,
				"type",
				"role",
				created_at,
				modified_at, visible) VALUES($1, $2, 1, 0, now(), now(), true);`
			fmt.Println("New record ID is:", newFileId)

			//userId when resource is url
			_, errInsert := db.Exec(insertUserFile, 1, newFileId)
			if errInsert != nil {
				panic(err)
			}
			providerUrl = fmt.Sprintf(
				"https://storage.googleapis.com/imgmdf/%d/%s",
				1,
				resHash)
			fmt.Println("providerUrl", providerUrl)
		}

		u, err := url.Parse(providerUrl)
		if err != nil {
			panic(err)
		}
		fmt.Print(providerUrl)

		buf, err := fetchImage(u)
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

			// 5. Perform transformations,
			// and save transformed image to blob and db
			newImage := resize.Resize(uint(width), uint(height), img, resize.Lanczos3)

			bufOut := new(bytes.Buffer)
			err = jpeg.Encode(bufOut, newImage, nil)
			sendBuf := bufOut.Bytes()
			//start save on google storage cloud
			ctx := context.Background()
			client, err := storage.NewClient(ctx, option.WithCredentialsFile("MyProject-89e0f34eb7a6.json"))

			if err != nil {
				log.Fatalf("Failed to create client: %v", err)
			}
			// Sets the name for the new bucket.
			bucketName := "imgmdf"
			// Creates a Bucket instance.
			bucket := client.Bucket(bucketName)
			obj := bucket.Object("1/" + orHash + "_/" + resModHash)
			wc := obj.NewWriter(ctx)

			if _, err = io.Copy(wc, bufOut); err != nil {
				panic(err)
			}
			if err := wc.Close(); err != nil {
				panic(err)
			}

			acl := obj.ACL()
			if err := acl.Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
				panic(err)
			}
			//end save on google storage cloud
			//save metadata in db
			sqlStatement := `
			INSERT INTO public.file
				(created_at, modified_at, last_opened, guid, hash, "index", parent_id, visible, "name", description, "type", allowed_operations, performed_operations)
				VALUES(now(), now(), now(), $1, $2, 0, null, true, $3, '', 1, '', $4) RETURNING id;`
			fileGuid, err := uuid.NewRandom()
			fmt.Print("fileGuid", fileGuid)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("%s", fileGuid)
			var newFileId int
			errIn := db.QueryRow(sqlStatement, fileGuid, resModHash, paramResource, resourceModifiers).Scan(&newFileId)
			if errIn != nil {
				log.Fatal(errIn)
			}

			insertUserFile := `
			INSERT INTO public.user_file
			(user_id,
			 	file_id,
				"type",
			 	"role",
				created_at,
			  modified_at, visible) VALUES($1, $2, 1, 0, now(), now(), true);`
			fmt.Println("New record ID is:", newFileId)

			//userId when resource is url
			_, errInsert := db.Exec(insertUserFile, 1, newFileId)
			if errInsert != nil {
				panic(err)
			}

			// 6. Return edited image
			w.Header().Set("Content-Length", strconv.Itoa(len(sendBuf)))
			w.Header().Set("Content-Type", "image/jpeg")
			w.Write(sendBuf)
			db.Close()
			return
		}
	}
}

func fetchImage(url *url.URL) ([]byte, error) {
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

func main() {
	port := os.Getenv("PORT")
	log.Println(port)
	if port == "" {
		port = "3001"
		log.Fatal("$PORT must be set")
	}
	// http.HandleFunc("/", handler)
	// log.Fatal(http.ListenAndServe(":"+port, nil))
	router := httprouter.New()
	router.GET("/user/:user/modifiers/:modifiers/resource/*resource", handler)
	router.GET("/user/:user/resource/*resource", handler)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
