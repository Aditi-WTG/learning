package main

import (
	"encoding/json"
	"os"
)

func Load[T any](fileName string) ([]T, error) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	var items []T
	err = json.Unmarshal(data, &items)
	if err != nil {
		return nil, err
	}

	return items, nil
}
