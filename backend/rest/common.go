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
		fmt.Printf("fail with error: %v\n", err)
		http.Error(w, err.Error(), statusCode)
	} else {
		fmt.Printf("fail with json: %v\n", err)
		w.WriteHeader(statusCode)
		fmt.Fprintln(w, string(responseJson))
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
		statusCode = v.StatusCode
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

func headers(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%s: %s\n", r.Method, r.RequestURI)
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func Handle(f func(w http.ResponseWriter, r *http.Request) (interface{}, error)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := f(w, r)
		if err != nil {
			fail(w, err)
		} else {
			success(w, data)
		}
	}
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("404 %s: %s\n", r.Method, r.RequestURI)
	http.NotFound(w, r)
}
