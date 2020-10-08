package model

type CustomDomainAcquireRequest struct {
	Domain   *string `json:"domain,omitempty"`
	Password *string `json:"password,omitempty"`
	Email    *string `json:"email,omitempty"`
}
