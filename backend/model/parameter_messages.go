package model

type ParameterMessages struct {
	Parameter string   `json:"parameter,omitempty"`
	Messages  []string `json:"messages,omitempty"`
}
