package imageserver

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/lib/pq"
)

func getUsage(userName string) map[int]int {
	db, err := sql.Open("postgres", connectionString)
	rows, err := db.Query(`select "type", count(Id) from logs where user_id = (select id from "users" where username = $1) group by "type"`,
		userName)
	if err, ok := err.(*pq.Error); ok {
		fmt.Println("pq error:", err.Code.Name())
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
	db.Close()
	return res
}

func checkLimit(operationType int, stats map[int]int) (bool, error) {
	limits := map[int]int{
		servedFromCache:            150 * 1000,
		servedOriginalImage:        150 * 1000,
		downloadSaveResourceInBlob: 150 * 1000,
		performOperations:          150 * 1000,
	}
	if limit, ok := limits[operationType]; ok {
		if usage, ok := stats[operationType]; ok {
			if usage >= limit {
				fmt.Printf("Limit reached limit: %d, usage: %d", limit, usage)
				return false, fmt.Errorf("Error.")
			}
			return true, nil
		}
		return true, nil
	}

	return false, fmt.Errorf("Error.")
}
