package main

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
	rows, err := db.Query(`select f.id, f.hash, uf.user_id, f.type, f.allowed_operations from file f
join user_file uf on uf.file_id = f.id
where (f.hash = $1 and f.type = 0) or f.hash = $2 and uf.user_id = (select id from "user" where username = $3)`,
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
	INSERT INTO public.file
		(created_at,
			modified_at,
			last_opened,
			guid,
			hash,
			"index",
			parent_id,
			visible,
			"name",
			description,
			"type",
			allowed_operations,
			performed_operations)
		VALUES(now(), now(), now(), $1, $2, 0, null, true, $3, '', $4, '', $5) RETURNING id;`
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
	errIn := db.QueryRow(
		sqlStatement,
		newFile.GUID,
		newFile.Hash,
		newFile.Name,
		newFile.Type,
		newFile.PerformedOperations,
	).Scan(&newFileId)

	if errIn != nil {
		return fileEntity{}, errIn
	}

	insertUserFile := `
	INSERT INTO public.user_file
	(user_id,
		file_id,
		"type",
		"role",
		created_at,
		modified_at, visible) VALUES((select id from "user" where username = $1), $2, 1, 0, now(), now(), true) RETURNING user_id;`
	var newFileUserId int
	errInsert := db.QueryRow(
		insertUserFile,
		newFile.UserName,
		newFileId).Scan(&newFileUserId)
	if errInsert != nil {
		return fileEntity{}, errInsert
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
	rows, err := db.Query(`select id, domain from "domain" where type = $2 and user_id = (select id from "user" where username = $1)`,
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
