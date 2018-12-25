package models

import (
	"fmt"
	u "imgserver/api/utils"
	"log"

	"github.com/jinzhu/gorm"
)

type Log struct {
	gorm.Model
	Body   string `json:"name"`
	Type   uint   `json:"type"`
	UserId uint   `json:"user_id"`
	FileId uint   `json:"file_id"`
}

/*
	0 served from cache
	1 served original image
	2 download resource and save in blob
	3 performOperations
*/

/*
 This struct function validate the required parameters sent through the http request body

returns message and true if the requirement is met
*/
func (log *Log) Validate() (map[string]interface{}, bool) {

	if log.FileId <= 0 {
		return u.Message(false, "File is not recognized"), false
	}

	if log.UserId <= 0 {
		return u.Message(false, "User is not recognized"), false
	}

	//All the required parameters are present
	return u.Message(true, "success"), true
}

func (log *Log) Create() map[string]interface{} {

	if resp, ok := log.Validate(); !ok {
		return resp
	}

	GetDB().Create(log)

	resp := u.Message(true, "success")
	resp["log"] = log
	return resp
}

func GetLog(id uint) *Log {

	log := &Log{}
	err := GetDB().Table("logs").Where("id = ?", id).First(log).Error
	if err != nil {
		return nil
	}
	return log
}

func GetLogs(user uint) []*Log {
	logs := make([]*Log, 0)
	err := GetDB().Table("logs").Where("user_id = ?", user).Find(&logs).Error
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return logs
}

func GetUsage(userId string) map[int]int {
	res := make(map[int]int)
	rows, err := GetDB().Table("logs").Where(`user_id = (select id from "users" where username = ?)`, userId).Select(`"type" as type, count(Id) as total`).Group("type").Rows()
	if err != nil {
		return nil
	}
	for rows.Next() {
		var (
			requestType  int
			requestCount int
		)
		err = rows.Scan(&requestType, &requestCount)
		if err != nil {
			log.Fatal(err)
		}
		res[requestType] = requestCount
	}
	return res
}
