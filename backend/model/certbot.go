package model

type CertbotPresentRequest struct {
	Token string `json:"token,omitempty"`
	Fqdn  string `json:"fqdn,omitempty"`
	Value string `json:"value,omitempty"`
}

type CertbotCleanUpRequest struct {
	Token string `json:"token,omitempty"`
	Fqdn  string `json:"fqdn,omitempty"`
	Value string `json:"value,omitempty"`
}
