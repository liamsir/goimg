package imageserver

import (
	"fmt"
	"imgserver/api/models"
	"net/url"
	"strconv"

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

	/*
		Load all file hashes in cache
	*/
	fileHashes, err := getFilesForAllUsers()
	if err != nil {
		panic(err)
	}

	pipe := client.Pipeline()
	//loadArray(fileHashes, &pipe)
	for i, elem := range fileHashes {
		fmt.Println(elem)
		// err := pipe.Set(elem.Hash, elem.Usage, 0).Err()
		_, err := pipe.HMSet(elem.Hash, map[string]interface{}{
			"FileId": elem.FileId,
			"UserId": elem.UserId,
			"Status": elem.Status,
		}).Result()

		if err != nil {
			fmt.Println(err)
			panic(err)
		}
		if i%1000 == 0 {
			pipe.Exec()
		}
	}
	pipe.Exec()
	/*
		Load all domains in cache
	*/
	allowedDomains, err := getDomainsForAllUsers()
	if err != nil {
		panic(err)
	}
	loadArray(allowedDomains, &pipe)

	/*
		Load usage for all users
	*/
	usage, err := getUsageForAllUsers()
	if err != nil {
		panic(err)
	}
	for i, elem := range usage {
		err := pipe.Set(elem.Hash, elem.Usage, 0).Err()
		if err != nil {
			panic(err)
		}
		if i%1000 == 0 {
			pipe.Exec()
		}
	}
	pipe.Exec()
}

func loadArray(list []string, pipe *redis.Pipeliner) {
	for i := 0; i < len(list); i++ {
		err := (*pipe).Set(list[i], true, 0).Err()
		if err != nil {
			panic(err)
		}

		if i%1000 == 0 {
			(*pipe).Exec()
		}
	}
	(*pipe).Exec()
}

func GetClient() *redis.Client {
	return client
}

func CheckOrigin(params CheckOriginParams) error {
	if params.Request.Referer() == "" {
		return nil
	}

	u, err := url.Parse(params.Request.Referer())
	if err != nil {
		return fmt.Errorf("Failed to parse requeset referer.")
	}
	key := fmt.Sprintf("_domain_%s%s", params.UserName, u.Scheme+"://"+u.Hostname()+"1")
	val, err := client.Get(key).Result()
	if err != nil {
		fmt.Println(key)

		return fmt.Errorf("Domain not allowed.")
	}

	if val == "1" {
		return nil
	}

	return fmt.Errorf("Domain not allowed.")
}

func fileStatus(userName string, version string) (int, bool) {
	key := fmt.Sprintf("_file_%s%s", userName, version)

	vals, err := client.HMGet(key, "Status").Result()
	if err != nil {
		panic(err)
	}
	if len(vals) == 0 || vals[0] == nil {
		return 0, false
	}
	res := vals[0].(string)
	i, err := strconv.Atoi(res)
	return i, true
}

func fileMeta(userName string, version string) (models.UserFile, bool) {
	key := fmt.Sprintf("_file_%s%s", userName, version)

	vals, err := client.HMGet(key, "Status", "UserId", "FileId").Result()
	if err != nil {
		return models.UserFile{}, false
	}

	status := vals[0].(string)
	userId := vals[1].(string)
	fieldId := vals[2].(string)

	istatus, err := strconv.Atoi(status)
	iuserid, err := strconv.Atoi(userId)
	ifieldid, err := strconv.Atoi(fieldId)

	return models.UserFile{Status: int32(istatus), UserId: int32(iuserid), FileId: int32(ifieldid)}, true
}

func setFileStatus(userName string, version string, status int) error {
	key := fmt.Sprintf("_file_%s%s", userName, version)
	err := client.HMSet(key, map[string]interface{}{
		"FileId": 0,
		"UserId": 0,
		"Status": status,
	}).Err()
	if err != nil {
		return err
	}
	return nil
}

func UpdateFileStatus(userName string, version string, status int, fileId int, userId int) error {
	key := fmt.Sprintf("_file_%s%s", userName, version)
	err := client.HMSet(key, map[string]interface{}{
		"FileId": fileId,
		"UserId": userId,
		"Status": status,
	}).Err()
	if err != nil {
		return err
	}
	return nil
}

func incrUsage(userName string, operation int) error {
	key := fmt.Sprintf("_usage_%s%d", userName, operation)
	err := client.Incr(key).Err()
	if err != nil {
		return err
	}
	return nil
}

func GetUsage(userName string, t int) (string, error) {
	key := fmt.Sprintf("_usage_%s%d", userName, t)
	val, err := client.Get(key).Result()
	return val, err
}
