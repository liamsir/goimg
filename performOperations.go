package main

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"net/http"
	"strconv"
	"strings"

	"github.com/nfnt/resize"
)

type performOperationsParam struct {
	paramModifiers string
	resourceUrl    string
	userName       string
	resource       string
}

func performOperationsAndWriteImageToRequest(params performOperationsParam, w http.ResponseWriter) (bool, error) {

	buf, err := fetchImage(params.resourceUrl)

	if err != nil {
		fmt.Println("failed to fetch image")
		return false, err
	}

	fileReader := bytes.NewReader(buf)
	img, err := jpeg.Decode(fileReader)

	if err != nil {
		fmt.Println("failed to decode image represented as bytes")
		return false, err
	}

	modifiers := strings.Split(params.paramModifiers, "_")
	if len(modifiers) > 0 {
		widthStr := modifiers[1]
		heightStr := modifiers[2]
		width, e := strconv.ParseUint(strings.Replace(widthStr, "w", "", -1), 10, 32)
		height, e := strconv.ParseUint(strings.Replace(heightStr, "h", "", -1), 10, 32)

		if e != nil {
			return false, e
		}
		// 5. Perform transformations, and save transformed image to blob and db
		newImage := resize.Resize(uint(width), uint(height), img, resize.Lanczos3)
		bufOut := new(bytes.Buffer)
		err = jpeg.Encode(bufOut, newImage, nil)
		sendBuf := bufOut.Bytes()
		fmt.Println("length of sendBuf", len(sendBuf))
		//save metadata in db
		newFile, err := saveFileEntity(fileEntity{
			Type:                1,
			UserName:            params.userName,
			Name:                params.resource,
			PerformedOperations: params.paramModifiers,
		})

		if err != nil {
			fmt.Println("failed to save file entity", err.Error())
			return false, err
		}
		resourceHash := hash(params.resource)
		fileObject := fileObject{
			Body: sendBuf,
			Name: fmt.Sprintf("%d/%d_/%s", newFile.UserId, resourceHash, newFile.Hash),
		}
		saveObject(fileObject)
		// 6. Return edited image
		w.Header().Set("Content-Length", strconv.Itoa(len(sendBuf)))
		w.Header().Set("Content-Type", "image/jpeg")
		w.Write(sendBuf)
		return true, nil
	}
	return false, fmt.Errorf("Error.")
}
