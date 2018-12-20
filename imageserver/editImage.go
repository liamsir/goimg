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

type performOperationsParamTest struct {
	paramModifiers string
	resourceUrl    string
	userName       string
	resource       string
}

/*
Resize
Parameters
Size
Width and height to set the image to.
Large 1920 1920
Medium 500 500
Thumb 150 150
*/
type imageOperation struct {
	name  string
	value map[string]int
}

func editImage(params performOperationsParam, w http.ResponseWriter, usageStats map[int]int) (bool, error) {

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

	modifiers, err := parseModifiers(params.paramModifiers)
	if err != nil {
		return false, fmt.Errorf("Error.")
	}
	fmt.Println(modifiers)
	if len(modifiers) > 0 {
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

		if err != nil {
			return false, err
		}
		//newImage := resize.Thumbnail(100, 1, img, resize.Lanczos3)
		bufOut := new(bytes.Buffer)
		err = jpeg.Encode(bufOut, img, nil)
		sendBuf := bufOut.Bytes()
		fmt.Println("length of sendBuf", len(sendBuf))

		// 6. Return edited image
		w.Header().Set("Content-Length", strconv.Itoa(len(sendBuf)))
		w.Header().Set("Content-Type", "image/jpeg")
		w.Write(sendBuf)
		return true, nil
	}
	return false, fmt.Errorf("Error.")
}
