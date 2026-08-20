package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

func ScanJSON(value []byte, target any) error {
	if len(value) == 0 {
		value = []byte("null")
	}
	return json.Unmarshal(value, target)
}

func JSONString(value any) (string, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

type NullString struct {
	sql.NullString
}

func NullableString(value string) NullString {
	if value == "" {
		return NullString{NullString: sql.NullString{Valid: false}}
	}
	return NullString{NullString: sql.NullString{String: value, Valid: true}}
}

func (n NullString) StringOrEmpty() string {
	if n.Valid {
		return n.String
	}
	return ""
}

func MustString(result sql.Result, err error) error {
	return fmt.Errorf("unused result: %v %w", result, err)
}
