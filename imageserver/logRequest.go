package imageserver

import (
	"database/sql"
)

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
