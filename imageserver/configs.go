package imageserver

var connectionString string = "postgres://jxbnzxtecqvcsv:9f603a3b7a60b5583f668fa2cf0ab0badd2c8f9dbacc73564cb1e9ee45241312@ec2-54-246-85-234.eu-west-1.compute.amazonaws.com:5432/dag2mo4a48vlb3"

// var connectionString string = "postgresql://rootgoimgserver:9f603a3b7a60b5583f66@goimgserver.cglo5epcd2hd.us-east-2.rds.amazonaws.com:5432/goimgserver"
var storageBucketUrl string = "https://storage.googleapis.com/imgmdf"
var storageBucketName string = "imgmdf"

// var redisHost string = "18.220.240.4:6379"
// var redisPwd string = "D+zs1DGBjBcO8evHrTvaGYyygVr+hmIDFCn21xTooF5X7Rioc44Ay/av4/mBx2r6J6NqC68CMvZSZ1vB"

var redisHost string = "127.0.0.1:6379"
var redisPwd string = ""
