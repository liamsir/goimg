package imageserver

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type fileEntity struct {
	Id                  int
	Name                string
	Hash                string
	UserId              int
	MasterId            int
	Type                int32
	AllowedOperations   string
	PerformedOperations string
	UserName            string
	GUID                string
	ResourceURL         string
}

type getFileParams struct {
	userName  string
	modifiers string
	resource  string
}

func getResourceInfo(params getFileParams) map[int32]fileEntity {

	resourceModifiers := params.resource + params.modifiers
	resourceNameModifiersHash := fmt.Sprint(hash(resourceModifiers))
	resourceNameHash := fmt.Sprint(hash(params.resource))
	userName := params.userName

	// 1. Check if resource exists in cache for given parameters
	db, err := sql.Open("postgres", connectionString)
	rows, err := db.Query(`select id, hash, user_id, type, allowed_operations from files
		where (hash = $1 and type = 0) or hash = $2 and user_id = (select id from "users" where username = $3)`,
		resourceNameHash, resourceNameModifiersHash, userName)
	if _, ok := err.(*pq.Error); ok {
		defer db.Close()
		return nil
	}
	res := make(map[int32]fileEntity)
	for rows.Next() {
		var (
			file              fileEntity
			allowedOperations sql.NullString
		)
		err = rows.Scan(&file.Id, &file.Hash, &file.UserId, &file.Type, &allowedOperations)
		if err != nil {
			log.Fatal(err)
		}
		if allowedOperations.Valid {
			file.AllowedOperations = allowedOperations.String
		}
		res[file.Type] = file
	}
	defer db.Close()
	return res
}

func saveFileEntity(newFile fileEntity) (fileEntity, error) {
	sqlStatement := `
	INSERT INTO public.files
		(created_at,
			guid,
			hash,
			"name",
			"type",
			allowed_operations,
			operations,
			user_id,
			master_id
		)
		VALUES(now(),  $1, $2, $3, $4, '', $5, (select id from "users" where username = $6), $7) RETURNING id, user_id;`
	fileGuid, err := uuid.NewRandom()

	if err != nil {
		return fileEntity{}, fmt.Errorf("failed to generate new guid")
	}
	newFileHash := hash(fmt.Sprintf("%s%s", newFile.Name, newFile.PerformedOperations))
	newFile.GUID = fileGuid.String()
	newFile.Hash = fmt.Sprintf("%d", newFileHash)

	db, err := sql.Open("postgres", connectionString)

	if err != nil {
		defer db.Close()
		return fileEntity{}, err
	}

	var newFileId int
	var newFileUserId int
	errIn := db.QueryRow(
		sqlStatement,
		newFile.GUID,
		newFile.Hash,
		newFile.Name,
		newFile.Type,
		newFile.PerformedOperations,
		newFile.UserName,
		newFile.MasterId,
	).Scan(&newFileId, &newFileUserId)

	if errIn != nil {
		return fileEntity{}, errIn
	}

	newFile.Id = newFileId
	newFile.UserId = newFileUserId
	defer db.Close()
	return newFile, nil
}

type domainEntity struct {
	Id     int
	Domain string
}

func getAllowedDomains(userName string, checkType int32) map[string]bool {

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		fmt.Println("Failed to open a connection!")
	}
	rows, err := db.Query(`select id, name from "domains" where type = $2 and user_id = (select id from "users" where username = $1)`,
		userName, checkType)
	if _, ok := err.(*pq.Error); ok {
		defer db.Close()
		return nil
	}
	res := make(map[string]bool)
	for rows.Next() {
		var (
			userDomain domainEntity
		)
		err = rows.Scan(&userDomain.Id, &userDomain.Domain)
		if err != nil {
			log.Fatal(err)
		}
		res[userDomain.Domain] = true
	}
	defer db.Close()
	return res
}

const (
	servedFromCache            = 0
	servedOriginalImage        = 1
	downloadSaveResourceInBlob = 2
	performOperations          = 3
	uploadImage                = 4
)

type requestEntity struct {
	Id     int
	Body   string
	UserId int
	FileId int
	Type   int32
	// 0 served from cache
	// 1 served original image
	// 2 download resource and save in blob
	// 3 performOperations
}

func logRequest(requestInfo requestEntity) (requestEntity, error) {

	sqlStatement := `INSERT INTO public.logs (created_at, user_id, file_id, body, "type")
  VALUES(now(), $1 , $2, $3, $4) RETURNING ID;`

	db, err := sql.Open("postgres", connectionString)

	if err != nil {
		return requestEntity{}, err
	}

	var newLogId int
	errIn := db.QueryRow(
		sqlStatement,
		requestInfo.UserId,
		requestInfo.FileId,
		requestInfo.Body,
		requestInfo.Type,
	).Scan(&newLogId)

	if errIn != nil {
		defer db.Close()
		return requestEntity{}, errIn
	}
	requestInfo.Id = newLogId
	defer db.Close()

	return requestInfo, nil
}

func getUsage(userName string) map[int]int {
	db, err := sql.Open("postgres", connectionString)
	rows, err := db.Query(`select "type", count(Id) from logs where user_id = (select id from "users" where username = $1) group by "type"`,
		userName)
	if _, ok := err.(*pq.Error); ok {
		defer db.Close()
		return nil
	}
	res := make(map[int]int)
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
	defer db.Close()
	return res
}
