package model


type SendLogsRequest struct {
	Token          string `json:"token"`
	Data           string `json:"data"`
	IncludeSupport bool   `json:"include_support"`
}