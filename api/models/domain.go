package models

import (
	u "imgserver/api/utils"

	"github.com/jinzhu/gorm"
)

type Domain struct {
	gorm.Model
	Name   string `json:"name"`
	Type   uint   `json:"type"`
	UserId uint   `json:"user_id"` //The user that this contact belongs to
}

/*
 This struct function validate the required parameters sent through the http request body

returns message and true if the requirement is met
*/
func (domain *Domain) Validate() (map[string]interface{}, bool) {

	if domain.Name == "" {
		return u.Message(false, "Domain name should be on the payload"), false
	}

	if domain.UserId <= 0 {
		return u.Message(false, "User is not recognized"), false
	}

	//All the required parameters are present
	return u.Message(true, "success"), true
}

func (domain *Domain) Create() map[string]interface{} {

	if resp, ok := domain.Validate(); !ok {
		return resp
	}

	GetDB().Create(domain)

	resp := u.Message(true, "success")
	resp["domain"] = domain
	return resp
}

func GetDomain(id uint) *Domain {

	contact := &Domain{}
	err := GetDB().Table("domains").Where("id = ?", id).First(contact).Error
	if err != nil {
		return nil
	}
	return contact
}
