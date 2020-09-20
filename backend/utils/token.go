package utils

import (
	"github.com/google/uuid"
)

func Uuid() string {
	id := uuid.New()
	return id.String()
}
