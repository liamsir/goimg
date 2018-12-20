package controllers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"imgserver/api/models"
	u "imgserver/api/utils"
	"imgserver/imageserver"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

var CreateFile = func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	user := r.Context().Value("user").(uint) //Grab the id of the user that send the request
	file := &models.File{}

	err := json.NewDecoder(r.Body).Decode(file)
	if err != nil {
		u.Respond(w, u.Message(false, "Error while decoding request body"))
		return
	}

	file.UserId = user
	resp := file.Create()
	u.Respond(w, resp)
}

var GetFilesFor = func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		//The passed path parameter is not an integer
		u.Respond(w, u.Message(false, "There was an error in your request"))
		return
	}

	data := models.GetFilesFor(uint(id))
	resp := u.Message(true, "success")
	resp["data"] = data
	u.Respond(w, resp)
}

var UploadImage = func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	signature := ps.ByName("signature")
	fileName := ps.ByName("fileName")
	userName := ps.ByName("user")

	res, err := GetImage(r)
	if err != nil {
		u.Respond(w, u.Message(false, "There was an error in your request"))
		return
	}

	user := models.GetUserWithUsername(userName)
	if user == nil {
		u.Respond(w, u.Message(false, "There was an error in your request"))
		return
	}
	_, errValidating := validateSignature(ValidateSignatureParams{
		SecretKey: user.SecretKey,
		Signature: signature,
		FileName:  fileName,
		Image:     res,
		Username:  userName,
	})
	if errValidating != nil {
		u.Respond(w, u.Message(false, "There was an error in your request while validating signature."))
		return
	}
	//return uploaded image url

	// resp["signKey"] = signKey

	// upload image
	fileHash := fmt.Sprintf("%d", hash(fileName))
	var fileObject = imageserver.FileObject{
		Body: res,
		Name: fmt.Sprintf("%d/%s", user.ID, fileHash),
	}
	imageserver.SaveObject(fileObject)
	// return image url

	const serverUrl = "http://localhost:3000"
	imageUrl := fmt.Sprintf("%s/user/%s/resource/%s",
		serverUrl,
		userName,
		fileName)

	fileGuid, err := uuid.NewRandom()

	if err != nil {
		u.Respond(w, u.Message(false, "There was an error in your request while validating signature."))
		return
	}
	newFile := models.File{}
	newFile.UserId = user.ID
	newFile.Guid = fileGuid.String()
	newFile.Hash = fileHash
	newFile.Name = fileName

	newFile.Create()

	newLog := models.Log{}
	newLog.Body = ""
	newLog.FileId = newFile.ID
	newLog.UserId = user.ID
	newLog.Type = 4
	newLog.Create()

	resp := u.Message(true, "success")
	resp["imageUrl"] = imageUrl

	u.Respond(w, resp)
}

func GetImage(r *http.Request) ([]byte, error) {
	if isFormBody(r) {
		return readFormBody(r)
	}
	return readRawBody(r)
}
func isFormBody(r *http.Request) bool {
	return strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/")
}

const formFieldName = "file"
const maxMemory int64 = 1024 * 1024 * 64

func readRawBody(r *http.Request) ([]byte, error) {
	return ioutil.ReadAll(r.Body)
}
func readFormBody(r *http.Request) ([]byte, error) {
	err := r.ParseMultipartForm(maxMemory)
	if err != nil {
		return nil, err
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buf, err := ioutil.ReadAll(file)
	if len(buf) == 0 {
		err = fmt.Errorf("Error")
	}

	return buf, err
}

type ValidateSignatureParams struct {
	SecretKey string
	Signature string
	Username  string
	FileName  string
	Image     []byte
}

func validateSignature(params ValidateSignatureParams) (bool, error) {

	mimeType := http.DetectContentType(params.Image)

	h := hmac.New(sha256.New, []byte(params.SecretKey))
	h.Write([]byte(params.Username))
	h.Write([]byte(params.FileName))
	h.Write([]byte(mimeType))
	h.Write([]byte(fmt.Sprintf("%d", len(params.Image))))

	expectedSign := h.Sum(nil)
	urlSign, err := base64.RawURLEncoding.DecodeString(params.Signature)

	if err != nil {
		return false, err
	}

	if hmac.Equal(urlSign, expectedSign) == false {
		return false, fmt.Errorf("Error.")
	}
	return true, nil
}

var SignUrl = func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	signUrl := &models.SignUrlViewModel{}

	err := json.NewDecoder(r.Body).Decode(signUrl)
	if err != nil {
		u.Respond(w, u.Message(false, "Error while decoding request body"))
		return
	}
	signUrl.UserName = ps.ByName("user")

	if resp, ok := signUrl.Validate(); !ok {
		u.Respond(w, resp)
		return
	}

	signData, err := signUrlHelper(signUrl)

	if err != nil {
		u.Respond(w, u.Message(false, "Error while decoding request body"))
		return
	}

	user := models.GetUserWith(signUrl.UserName, signUrl.SecretKey)

	if user == nil {
		u.Respond(w, u.Message(false, "Invalid username or secret key"))
		return
	}
	fmt.Println(user)
	var data = make(map[string]interface{})
	data["uploadUrl"] = fmt.Sprintf("http://localhost:3000/user/%s/upload/%s/file/%s",
		user.Username,
		signData,
		signUrl.Image.Name)

	resp := u.Message(true, "success")
	resp["data"] = data
	u.Respond(w, resp)
}

func signUrlHelper(signUrl *models.SignUrlViewModel) (string, error) {
	h := hmac.New(sha256.New, []byte(signUrl.SecretKey))
	h.Write([]byte(signUrl.UserName))
	h.Write([]byte(signUrl.Image.Name))
	h.Write([]byte(signUrl.Image.ContentType))
	h.Write([]byte(fmt.Sprintf("%d", signUrl.Image.ContentLength)))
	buf := h.Sum(nil)
	signature := base64.RawURLEncoding.EncodeToString(buf)
	return signature, nil
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
