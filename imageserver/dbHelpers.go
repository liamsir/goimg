package imageserver

import (
	"fmt"
	"imgserver/api/models"
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

func getResourceInfo(params getFileParams) map[uint]models.File {

	master := fmt.Sprint(hash(params.resource))
	version := fmt.Sprint(hash(params.resource + params.modifiers))
	userName := params.userName

	resp := models.GetFilesForHash(master, version, userName)
	res := make(map[uint]models.File)
	for _, element := range resp {
		fmt.Println("elem", element)
		res[element.Type] = *element
	}
	return res
}

func saveFileEntity(newFile fileEntity) (fileEntity, error) {

	user := models.GetUserWithUsername(newFile.UserName)

	file := models.File{}

	newFileHash := hash(fmt.Sprintf("%s%s", newFile.Name, newFile.PerformedOperations))
	file.Hash = fmt.Sprintf("%d", newFileHash)
	file.Name = newFile.Name
	file.Type = uint(newFile.Type)
	file.Operations = newFile.PerformedOperations
	file.MasterId = uint(newFile.MasterId)
	file.UserId = user.ID

	res := file.Create()

	createdFile := res["file"].(*models.File)

	newFile.Id = int(createdFile.ID)
	newFile.UserId = int(createdFile.UserId)
	newFile.Hash = file.Hash

	return newFile, nil
}

type domainEntity struct {
	Id     int
	Domain string
}

func getAllowedDomains(userName string, checkType int32) map[string]bool {
	return models.GetDomainsForUserName(userName, checkType)
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
}

func logRequest(requestInfo requestEntity) error {

	log := models.Log{
		UserId: uint(requestInfo.UserId),
		FileId: uint(requestInfo.FileId),
		Body:   requestInfo.Body,
		Type:   uint(requestInfo.Type),
	}

	log.Create()
	return nil
}

func getUsage(userName string) map[int]int {
	return models.GetUsage(userName)
}
