package imageserver

import (
	"fmt"

	"github.com/go-redis/redis"
)

var client *redis.Client

func init() {
	client = redis.NewClient(&redis.Options{
		Addr:     redisHost,
		Password: redisPwd,
		DB:       0,
	})
	pong, err := client.Ping().Result()
	fmt.Println(pong, err)
	// value := make(map[string]interface{})
	// //value["servedFromCache"] = 15
	// value["servedOriginalImage"] = 55
	//
	// errd := client.HMSet("user_1", value).Err()
	// if errd != nil {
	// 	panic(err)
	// }
	// client.Pipeline().HSet("sdf", "f", "v")
	//client.Pipeline().Exec()
}

func GetClient() *redis.Client {
	return client
}
