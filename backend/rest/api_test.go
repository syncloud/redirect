package rest

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/syncloud/redirect/model"
	"testing"
)

func TestParameterErrorToError(t *testing.T) {
	err := &model.ParameterError{ParameterErrors: &[]model.ParameterMessages{{
		Parameter: "param", Messages: []string{"error"},
	}}}
	response, code := ErrorToResponse(err)
	assert.Equal(t, 400, code)
	assert.Equal(t, "param", (*response.ParametersMessages)[0].Parameter)
}

func TestServiceErrorToError(t *testing.T) {
	err := model.NewServiceError("error")
	response, code := ErrorToResponse(err)
	assert.Equal(t, 400, code)
	assert.Equal(t, "error", response.Message)
}

func TestErrorToError(t *testing.T) {
	err := fmt.Errorf("error")
	response, code := ErrorToResponse(err)
	assert.Equal(t, 500, code)
	assert.Equal(t, "error", response.Message)
}
