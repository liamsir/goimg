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
	if err, ok := err.(*pq.Error); ok {
		fmt.Println("pq error:", err.Code.Name())
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
	db.Close()
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
			user_id
		)
		VALUES(now(),  $1, $2, $3, $4, '', $5, (select id from "users" where username = $6)) RETURNING id, user_id;`
	fileGuid, err := uuid.NewRandom()

	if err != nil {
		return fileEntity{}, fmt.Errorf("failed to generate new guid")
	}
	newFileHash := hash(fmt.Sprintf("%s%s", newFile.Name, newFile.PerformedOperations))
	newFile.GUID = fileGuid.String()
	newFile.Hash = fmt.Sprintf("%d", newFileHash)

	db, err := sql.Open("postgres", connectionString)

	if err != nil {
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
	).Scan(&newFileId, &newFileUserId)

	if errIn != nil {
		return fileEntity{}, errIn
	}

	newFile.Id = newFileId
	newFile.UserId = newFileUserId
	db.Close()
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
	if err, ok := err.(*pq.Error); ok {
		fmt.Println("pq error:", err.Code.Name())
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
	db.Close()
	return res
}
