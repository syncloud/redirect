package rest

import (
	"encoding/json"
	"fmt"
	"github.com/syncloud/redirect/model"
	"net/http"
)

type DomainAcquireResponse struct {
	Success              bool   `json:"success"`
	DeprecatedUserDomain string `json:"user_domain,omitempty"`
	UpdateToken          string `json:"update_token,omitempty"`
}

type Response struct {
	Success            bool                       `json:"success"`
	Message            string                     `json:"message,omitempty"`
	Data               *interface{}               `json:"data,omitempty"`
	ParametersMessages *[]model.ParameterMessages `json:"parameters_messages,omitempty"`
}

func fail(w http.ResponseWriter, err error) {
	response, statusCode := ErrorToResponse(err)
	responseJson, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), statusCode)
	} else {
		http.Error(w, string(responseJson), statusCode)
	}
}

func ErrorToResponse(err error) (Response, int) {
	response := Response{Success: false, Message: "Unknown Error"}
	statusCode := 500
	switch v := err.(type) {
	case *model.ParameterError:
		response.ParametersMessages = v.ParameterErrors
		statusCode = 400
	case *model.ServiceError:
		statusCode = 400
	}
	response.Message = err.Error()
	return response, statusCode
}
func success(w http.ResponseWriter, data interface{}) {
	response := Response{
		Success: true,
		Data:    &data,
	}
	responseJson, err := json.Marshal(response)
	if err != nil {
		fail(w, err)
	} else {
		_, _ = fmt.Fprintf(w, string(responseJson))
	}
}

func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func Handle(f func(req *http.Request) (interface{}, error)) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		data, err := f(req)
		if err != nil {
			fail(w, err)
		} else {
			success(w, data)
		}
	}
}
