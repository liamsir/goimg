package models

import (
	"fmt"
	u "imgserver/api/utils"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

type File struct {
	gorm.Model
	Name              string `json:"name"`
	Hash              string `json:"hash"`
	UserId            uint   `json:"user_id"` //The user that this contact belongs to
	Type              uint   `json:"type"`
	Operations        string `json:"operations"`
	AllowedOperations string `json:"allowed_operations"`
	Guid              string `json:"guid"`
}
type ImageMeta struct {
	Name          string
	ContentType   string
	ContentLength int32
}
type SignUrlViewModel struct {
	UserName  string
	SecretKey string
	Image     ImageMeta
}

//Validate incoming user details...
func (signUrl *SignUrlViewModel) Validate() (map[string]interface{}, bool) {

	if signUrl.SecretKey == "" {
		return u.Message(false, "SecretKey is required."), false
	}

	if signUrl.UserName == "" {
		return u.Message(false, "UserName is required."), false
	}

	if signUrl.Image.Name == "" {
		return u.Message(false, "Image name is required."), false
	}

	if signUrl.Image.ContentLength <= 0 {
		return u.Message(false, "Image content length is required."), false
	}

	if signUrl.Image.ContentType == "" {
		return u.Message(false, "Image content type is required."), false
	}

	return u.Message(false, "Requirement passed"), true
}

/*
 This struct function validate the required parameters sent through the http request body

returns message and true if the requirement is met
*/
func (file *File) Validate() (map[string]interface{}, bool) {

	if file.Name == "" {
		return u.Message(false, "File name should be on the payload"), false
	}

	if file.Hash == "" {
		return u.Message(false, "Hash should be on the payload"), false
	}

	if file.UserId <= 0 {
		return u.Message(false, "User is not recognized"), false
	}

	//All the required parameters are present
	return u.Message(true, "success"), true
}

func (file *File) Create() map[string]interface{} {

	if resp, ok := file.Validate(); !ok {
		return resp
	}

	fileGuid, err := uuid.NewRandom()

	if err != nil {
		return u.Message(false, "File name should be on the payload")
	}
	file.Guid = fileGuid.String()

	GetDB().Create(file)

	resp := u.Message(true, "success")
	resp["file"] = file
	return resp
}

func GetFile(id uint) *File {

	file := &File{}
	err := GetDB().Table("files").Where("id = ?", id).First(file).Error
	if err != nil {
		return nil
	}
	return file
}

func GetFilesFor(user uint) []*File {
	files := make([]*File, 0)
	err := GetDB().Table("files").Where("user_id = ?", user).Find(&files).Error
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return files
}
