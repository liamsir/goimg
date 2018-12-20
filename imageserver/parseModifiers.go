package imageserver

import (
	"fmt"
	"strconv"
	"strings"
)

func parseModifiers(modifiers string) ([]imageOperation, error) {
	allowedOperations := map[string]bool{"resize": true, "crop": true}
	const MAX_ALLOWED_OPERATIONS_PER_REQUEST = 3
	modifiersList := []imageOperation{}

	operations := strings.Split(modifiers, "&")

	if len(operations) > MAX_ALLOWED_OPERATIONS_PER_REQUEST {
		return nil, fmt.Errorf("Error.")
	}

	for i := 0; i < len(operations); i += 1 {
		// fmt.Println(operations[i])
		operation := strings.Split(operations[i], "_")
		if len(operation) > 0 {
			operationName := operation[0]
			// fmt.Println(operationName)
			_, ok := allowedOperations[operationName]
			if !ok {
				return nil, fmt.Errorf("Operation not allowed.")
			}
			var newOperation = imageOperation{}
			newOperation.name = operationName

			if operationName == "resize" {
				if len(operation) == 3 || len(operation) == 4 {
					width, e := strconv.ParseUint(strings.Replace(operation[1], "w", "", -1), 10, 32)
					height, e := strconv.ParseUint(strings.Replace(operation[2], "h", "", -1), 10, 32)
					var resizeMode = 0
					if len(operation) == 4 {
						resizeModeStr := strings.Replace(operation[3], "mode", "", -1)
						if resizeModeStr == "Max" {
							resizeMode = 1
						}
					}
					if e != nil {
						return nil, fmt.Errorf("Operation not allowed.")
					}
					newOperation.value = map[string]int{}
					newOperation.value["width"] = int(width)
					newOperation.value["height"] = int(height)
					newOperation.value["mode"] = resizeMode
				} else if len(operation) == 2 {
					var (
						width  int = 0
						height int = 0
					)
					if operation[1] == "large" {
						width = 1920
						height = 1920
					} else if operation[1] == "medium" {
						width = 500
						height = 500
					} else if operation[1] == "thumb" {
						width = 150
						height = 150
					}
					newOperation.value = map[string]int{}
					newOperation.value["width"] = int(width)
					newOperation.value["height"] = int(height)
					newOperation.value["mode"] = 1
				}
			} else if operationName == "crop" {
				if len(operation) == 5 {

					left, e := strconv.ParseUint(strings.Replace(operation[1], "x", "", -1), 10, 32)
					top, e := strconv.ParseUint(strings.Replace(operation[2], "y", "", -1), 10, 32)
					right, e := strconv.ParseUint(strings.Replace(operation[3], "w", "", -1), 10, 32)
					bottom, e := strconv.ParseUint(strings.Replace(operation[4], "h", "", -1), 10, 32)

					if e != nil {
						return nil, fmt.Errorf("Operation not allowed.")
					}
					newOperation.value = map[string]int{}
					newOperation.value["left"] = int(left)
					newOperation.value["top"] = int(top)
					newOperation.value["right"] = int(right)
					newOperation.value["bottom"] = int(bottom)
				}
			}
			modifiersList = append(modifiersList, newOperation)
		}
	}

	return modifiersList, nil
}
