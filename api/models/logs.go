package models

import (
	"fmt"
	u "imgserver/api/utils"
	"log"
	"time"

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

func GetLogs(user uint, page uint, start time.Time, end time.Time) []*Log {

	limit := uint(20)
	offset := (page - 1) * limit

	logs := make([]*Log, 0)
	err := GetDB().Table("logs").Order("created_at desc").Offset(offset).Limit(limit).Where("user_id = ? AND created_at >= ? AND created_at < ?", user, start, end).Find(&logs).Error
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

type UserUsage struct {
	UserName string
	Usage    map[int]int
}

func GetUsageForAllUsers() map[string]map[int]int {
	// res := make([]*UserUsage, 0)
	// _, err := GetDB().Table("logs").Select(`"type" as type, count(Id) as total`).Group("type").Rows()
	// if err != nil {
	// 	return nil
	// }
	// // for rows.Next() {
	// // 	usage := {}
	// // 	err = rows.Scan(&requestType, &requestCount)
	// // 	if err != nil {
	// // 		log.Fatal(err)
	// // 	}
	// // 	res[requestType] = requestCount
	// // }
	// return res
	return nil
}

type LogGroup struct {
	Type      uint      `json:"type"`
	Total     uint      `json:"total"`
	CreatedAt time.Time `json:"created_at"`
}

func GetReportFor(userId uint, start time.Time, end time.Time) []*LogGroup {
	res := make([]*LogGroup, 0)
	err := GetDB().Table("logs").Where(`user_id = ? AND created_at >= ? AND created_at::date <= ?`, userId, start, end).Select(`created_at::date as created_at, "type", count(*) as total`).Group("created_at::date, type").Order("created_at desc").Find(&res).Error
	if err != nil {
		return nil
	}
	return res
}

// func GetReportFor(userId uint, start time.Time, end time.Time) map[string]interface{} {
// 	res := make(map[string]interface{})
// 	rows, err := GetDB().Table("logs").Where(`user_id = ? AND created_at >= ? AND created_at < ?`, userId, start, end).Select(`"type", count(*) as total`).Group("type").Rows()
// 	if err != nil {
// 		return nil
// 	}
// 	for rows.Next() {
// 		var (
// 			requestType  int
// 			requestCount int
// 		)
// 		err = rows.Scan(&requestType, &requestCount)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		data := struct {
// 			RequestType  int
// 			RequestCount int
// 		}{
// 			requestType,
// 			requestCount,
// 		}
// 		res[fmt.Sprintf("%d", requestType)] = data
// 	}
// 	res["start"] = start
// 	res["end"] = end
// 	return res
// }
