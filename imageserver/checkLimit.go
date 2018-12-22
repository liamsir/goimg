package imageserver

import (
	"fmt"
)

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
