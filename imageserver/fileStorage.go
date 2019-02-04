package imageserver

import (
	"bytes"
	"context"
	"fmt"
	"imgserver/api/models"
	"io"
	"log"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type FileObject struct {
	Body []byte
	Name string
}

func SaveObject(object FileObject) {
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

	// acl := obj.ACL()
	// if err := acl.Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
	// 	panic(err)
	// }
}

func DeleteFiles(user *models.User, files []*models.File, masterIds map[int]*models.File) error {
	bucketName := "imgmdf"
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithCredentialsFile("MyProject-89e0f34eb7a6.json"))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	var d []string
	for i := 0; i < len(files); i += 1 {
		f := files[i]
		prefix := fmt.Sprintf("%s/%s", user.Username, f.Hash)

		if f.Type == 0 {
			q := storage.Query{Prefix: prefix}
			o := client.Bucket(bucketName).Objects(ctx, &q)
			for {
				ob, err := o.Next()
				if err == iterator.Done {
					break
				}
				if err != nil {
					return err
				}
				d = append(d, ob.Name)
			}
			d = append(d, prefix)
		} else if f.Type == 1 {
			if master, ok := masterIds[int(f.MasterId)]; ok {
				o := fmt.Sprintf("%s/%s_/%s", user.Username, master.Hash, f.Hash)
				d = append(d, o)
			}
		}
	}

	for j := 0; j < len(d); j += 1 {
		o := client.Bucket(bucketName).Object(d[j])
		if err := o.Delete(ctx); err != nil {
			//return err
		}
	}
	return nil
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
	contentType := ""
	buf, err := fetchImage(params.ResourceUrl, &contentType)

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
		Name: fmt.Sprintf("%s/%s", newFile.UserName, newFile.Hash),
	}
	SaveObject(fileObject)
	remoteResource :=
		fmt.Sprintf("%s/%s",
			storageBucketUrl,
			fmt.Sprintf("%s/%s", newFile.UserName, newFile.Hash),
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
