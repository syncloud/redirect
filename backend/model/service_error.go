package model

import "errors"

type ServiceError struct {
	InternalError error
	StatusCode    int
}

func (e *ServiceError) Error() string {
	return e.InternalError.Error()
}

func NewServiceError(message string) *ServiceError {
	return &ServiceError{InternalError: errors.New(message), StatusCode: 400}
}

func NewServiceErrorWithCode(message string, code int) *ServiceError {
	return &ServiceError{InternalError: errors.New(message), StatusCode: code}
}
