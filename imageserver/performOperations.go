package imageserver

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"net/http"
	"strconv"

	"github.com/nfnt/resize"
	"github.com/oliamb/cutter"
)

type performOperationsParam struct {
	paramModifiers string
	resourceUrl    string
	userName       string
	resource       string
	userId         int
	resourceHash   string
}

func performOperationsAndWriteImageToRequest(params performOperationsParam, w http.ResponseWriter, usageStats map[int]int) (bool, error) {
	_, err := checkLimit(performOperations, usageStats)
	if err != nil {
		return false, fmt.Errorf("Error.")
	}

	fileName := fmt.Sprintf("%d/%s",
		params.userId,
		params.resourceHash,
	)
	urlSigned, err := signUrl(fileName)
	if err != nil {
		return false, err
	}
	buf, err := fetchImage(urlSigned)
	fmt.Println("buf length", len(buf))

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

	modifiers, err := parseModifiers(params.paramModifiers)
	if err != nil {
		return false, fmt.Errorf("Error.")
	}
	fmt.Println(modifiers)
	if len(modifiers) > 0 {
		// 5. Perform transformations, and save transformed image to blob and db

		for i := 0; i < len(modifiers); i += 1 {
			modifier := modifiers[i]
			fmt.Println(modifier)
			if modifier.name == "resize" {
				if modifier.value["mode"] == 0 {
					img = resize.Resize(uint(modifier.value["width"]), uint(modifier.value["height"]), img, resize.Lanczos3)
				} else if modifier.value["mode"] == 1 {
					img = resize.Thumbnail(uint(modifier.value["width"]), uint(modifier.value["height"]), img, resize.Lanczos3)
				}
			}
			if modifier.name == "crop" {
				img, err = cutter.Crop(img, cutter.Config{
					Width:  modifier.value["right"],
					Height: modifier.value["bottom"],
					Anchor: image.Point{modifier.value["left"], modifier.value["top"]},
				})
				if err != nil {
					return false, err
				}
			}
		}

		bufOut := new(bytes.Buffer)
		err = jpeg.Encode(bufOut, img, nil)
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
		var fileObject = FileObject{
			Body: sendBuf,
			Name: fmt.Sprintf("%d/%d_/%s", newFile.UserId, resourceHash, newFile.Hash),
		}
		SaveObject(fileObject)
		// 6. Return edited image
		w.Header().Set("Content-Length", strconv.Itoa(len(sendBuf)))
		w.Header().Set("Content-Type", "image/jpeg")
		w.Write(sendBuf)
		return true, nil
	}
	return false, fmt.Errorf("Error.")
}
