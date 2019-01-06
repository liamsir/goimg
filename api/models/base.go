package models

import (
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var db *gorm.DB //database
var connectionString string = "postgres://jxbnzxtecqvcsv:9f603a3b7a60b5583f668fa2cf0ab0badd2c8f9dbacc73564cb1e9ee45241312@ec2-54-246-85-234.eu-west-1.compute.amazonaws.com:5432/dag2mo4a48vlb3"

func init() {

	conn, err := gorm.Open("postgres", connectionString)

	// conn.DB().SetMaxOpenConns(2)
	if err != nil {
		fmt.Print(err)
	}

	db = conn
	// db.LogMode(true)
	// db.Debug().AutoMigrate(&User{}, &File{}, &Log{}, &Domain{}) //Database migration
}

//returns a handle to the DB object
func GetDB() *gorm.DB {
	return db
}
