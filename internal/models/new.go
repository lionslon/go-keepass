package models

import (
	"encoding/json"
	"fmt"
	"io"
)

func NewDTO[T any](reader io.Reader) (T, error) {
	var dto T
	err := json.NewDecoder(reader).Decode(&dto)
	if err != nil {
		return dto, fmt.Errorf("cannot parse dto: cannot unmarshal request body: %w", err)
	}
	return dto, nil
}
