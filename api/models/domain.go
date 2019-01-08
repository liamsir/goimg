package models

import (
	u "imgserver/api/utils"
	"strconv"
	"strings"

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

	// Domain must be unique
	temp := &Domain{}

	//check for errors and duplicate domain
	err := GetDB().Table("domains").Where("user_id = ? AND name = ? AND type = ?", domain.UserId, domain.Name, domain.Type).First(temp).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return u.Message(false, "Connection error. Please retry"), false
	}
	if temp.Name != "" {
		return u.Message(false, "Domain address already exists."), false
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

func (domain *Domain) Update() map[string]interface{} {

	if resp, ok := domain.Validate(); !ok {
		return resp
	}
	domainToUpdate := Domain{}
	err := GetDB().Table("domains").Where("user_id = ? AND id = ?", domain.UserId, domain.ID).First(&domainToUpdate).Error
	if err != nil {
		return nil
	}
	domainToUpdate.Name = domain.Name
	domainToUpdate.Type = domain.Type

	GetDB().Save(&domainToUpdate)

	resp := u.Message(true, "success")
	resp["domain"] = domain
	return resp
}

func (domain *Domain) Patch() map[string]interface{} {

	if resp, ok := domain.Validate(); !ok {
		return resp
	}

	d := &Domain{}
	err := GetDB().Table("domains").Where("user_id = ? AND id = ?", domain.UserId, domain.ID).First(d).Error
	if err != nil {
		return nil
	}

	patchFields := map[string]interface{}{}
	if domain.Name != "" {
		patchFields["name"] = domain.Name
	}
	if domain.Type != d.Type {
		patchFields["type"] = domain.Type
	}

	db.Model(&d).Updates(patchFields)

	resp := u.Message(true, "success")
	resp["domain"] = domain
	return resp
}

func GetDomain(userId uint, id int) *Domain {

	contact := &Domain{}
	err := GetDB().Table("domains").Where("user_id = ? AND id = ?", userId, id).First(contact).Error
	if err != nil {
		return nil
	}
	return contact
}

func GetDomainsFor(user uint) []*Domain {
	domains := make([]*Domain, 0)
	err := GetDB().Table("domains").Where("user_id = ?", user).Find(&domains).Error
	if err != nil {
		return nil
	}
	return domains
}
func GetDomainsForUserName(username string, dType int32) map[string]bool {
	domains := make([]*Domain, 0)
	err := GetDB().Table("domains").Where(`user_id = (select id from "users" where username = ?) and type = ? `, username, dType).Find(&domains).Error
	if err != nil {
		return nil
	}
	res := make(map[string]bool)
	for _, element := range domains {
		res[element.Name] = true
	}

	return res
}
func DeleteDomain(userId int, domains string) (*Domain, error) {
	domainIds := []int{}
	for _, i := range strings.Split(domains, ",") {
		j, err := strconv.Atoi(i)
		if err != nil {
			panic(err)
		}
		domainIds = append(domainIds, j)
	}
	err := GetDB().Where("id IN(?) AND user_id = ?", domainIds, userId).Delete(&Domain{})
	if err != nil {
		return nil, err.Error
	}
	return &Domain{}, nil
}
