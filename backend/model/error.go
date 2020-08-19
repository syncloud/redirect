package model

type Error struct {
	StatusCode      int
	Error           error
	ParameterErrors *[]ParameterMessages
}

func ParametersError(parameterErrors *[]ParameterMessages) *Error {
	return &Error{400, nil, parameterErrors}
}

func ServiceError(error error) *Error {
	return &Error{400, error, nil}
}

func UnknownError(error error) *Error {
	return &Error{500, error, nil}
}
