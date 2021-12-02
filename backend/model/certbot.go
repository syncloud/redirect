package model

type CertbotPresentRequest struct {
	Token  string   `json:"token,omitempty"`
	Fqdn   string   `json:"fqdn,omitempty"`
	Values []string `json:"values,omitempty"`
}

type CertbotCleanUpRequest struct {
	Token string `json:"token,omitempty"`
	Fqdn  string `json:"fqdn,omitempty"`
}
