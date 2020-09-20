package model

type ServiceError struct {
	InternalError error
}

func (e *ServiceError) Error() string {
	return e.InternalError.Error()
}
