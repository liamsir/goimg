package imageserver

import (
	"fmt"
	"strconv"
	"strings"
)

const MAX_ALLOWED_OPERATIONS_PER_REQUEST = 3

func parseModifiers(modifiers string) ([]imageOperation, error) {

	modifiersList := []imageOperation{}
	allowedOperations := map[string]bool{"resize": true, "crop": true}
	allowedParams := map[string]bool{
		"resizewidth":  true,
		"resizeheight": true,
		"cropleft":     true,
		"croptop":      true,
		"cropright":    true,
		"cropbottom":   true,
	}

	if len(modifiers) == 0 {
		return nil, fmt.Errorf("Error.")
	}

	operations := strings.Split(modifiers, "&")

	if len(operations) == 0 || len(operations) > MAX_ALLOWED_OPERATIONS_PER_REQUEST {
		return nil, fmt.Errorf("Error.")
	}

	for i := 0; i < len(operations); i += 1 {
		operation := strings.Split(operations[i], "=")
		if len(operation) != 2 {
			return nil, fmt.Errorf("Error.")
		}
		operationName := operation[0]
		_, ok := allowedOperations[operationName]
		if !ok {
			return nil, fmt.Errorf("Operation not allowed.")
		}

		operationParams := strings.Split(operation[1], ",")
		var newOperation = imageOperation{}
		newOperation.name = operationName
		newOperation.value = map[string]int{}
		for j := 0; j < len(operationParams); j += 1 {
			operationPrm := strings.Split(operationParams[j], ":")
			if len(operationPrm) != 2 {
				return nil, fmt.Errorf("Error.")
			}
			v, e := strconv.ParseInt(string(operationPrm[1]), 10, 32)
			if e != nil {
				return nil, fmt.Errorf("Error.")
			}
			prmName := operationPrm[0]
			_, ok := allowedParams[operationName+prmName]
			if !ok {
				return nil, fmt.Errorf("Operation not allowed.")
			}
			newOperation.value[prmName] = int(v)
		}
		fmt.Println("newOperation", newOperation)
		modifiersList = append(modifiersList, newOperation)
	}

	if len(modifiers) > 0 && len(modifiersList) == 0 {
		return nil, fmt.Errorf("Failed to parse modifiers.")
	}

	return modifiersList, nil
}
