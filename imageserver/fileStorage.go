package imageserver

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

type FileObject struct {
	Body []byte
	Name string
}

func SaveObject(object FileObject) {
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

func downloadResourceAndSaveInBlob(params downloadAndSaveObjectParams, usageStats map[int]int) (fileEntity, error) {

	// validate mime types and size
	_, err := checkLimit(downloadSaveResourceInBlob, usageStats)
	if err != nil {
		return fileEntity{}, fmt.Errorf("Error.")
	}
	buf, err := fetchImage(params.ResourceUrl)

	if err != nil {
		return fileEntity{}, fmt.Errorf("Error.")
	}

	newFile, err := saveFileEntity(fileEntity{
		Type:     0,
		UserName: params.UserName,
		Name:     params.ResourceUrl,
	})
	if err != nil {
		fmt.Println("error happend")
		fmt.Printf("length %d \n", len(buf))
		return fileEntity{}, fmt.Errorf("Error.")
	}

	var fileObject = FileObject{
		Body: buf,
		Name: fmt.Sprintf("%d/%s", newFile.UserId, newFile.Hash),
	}

	SaveObject(fileObject)
	remoteResource :=
		fmt.Sprintf("%s/%s",
			storageBucketUrl,
			fmt.Sprintf("%d/%s", newFile.UserId, newFile.Hash),
		)
	newFile.ResourceURL = remoteResource

	logRequest(requestEntity{
		Body:   "",
		FileId: newFile.Id,
		UserId: newFile.UserId,
		Type:   2,
	})

	return newFile, nil
}
