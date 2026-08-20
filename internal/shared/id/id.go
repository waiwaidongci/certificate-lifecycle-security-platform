package id

import (
	"fmt"

	"github.com/google/uuid"
)

func New() string {
	return uuid.NewString()
}

func IsValid(value string) bool {
	_, err := uuid.Parse(value)
	return err == nil
}

func Must(value string) string {
	if !IsValid(value) {
		panic(fmt.Sprintf("invalid uuid %q", value))
	}
	return value
}
