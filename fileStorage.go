package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

type fileObject struct {
	Body []byte
	Name string
}

func saveObject(object fileObject) {
	fmt.Println("writing file...")
	r := bytes.NewReader(object.Body)
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithCredentialsFile("MyProject-89e0f34eb7a6.json"))

	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	// Sets the name for the new bucket.
	bucketName := "imgmdf"
	// Creates a Bucket instance.
	bucket := client.Bucket(bucketName)
	obj := bucket.Object(object.Name)
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
	fmt.Println("done writing file")
}

type downloadAndSaveObjectParams struct {
	ResourceUrl string
	UserName    string
}

func downloadResourceAndSaveInBlob(params downloadAndSaveObjectParams) (string, error) {

	buf, err := fetchImage(params.ResourceUrl)

	if err != nil {
		return "", fmt.Errorf("Error.")
	}

	newFile, err := saveFileEntity(fileEntity{
		Type:     0,
		UserName: params.UserName,
		Name:     params.ResourceUrl,
	})
	if err != nil {
		return "", fmt.Errorf("Error.")
	}

	fileObject := fileObject{
		Body: buf,
		Name: fmt.Sprintf("%d/%s", newFile.UserId, newFile.Hash),
	}

	saveObject(fileObject)
	remoteResource :=
		fmt.Sprintf("%s/%s",
			storageBucketUrl,
			fmt.Sprintf("%d/%s", newFile.UserId, newFile.Hash),
		)
	return remoteResource, nil
}

// buf, err := fetchImage(paramResource)
// if err != nil {
// 	panic(err)
// }
// // 		// Sets the name for the new bucket.
// saveObject(fileObject{Body: buf, Name: "1/" + "resHash"})

// providerUrl = fmt.Sprintf(
// 	"https://storage.googleapis.com/imgmdf/%d/%s",
// 	1,
// 	resHash)
// fmt.Println("providerUrl", providerUrl)
