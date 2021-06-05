package model

type DomainAvailabilityRequest struct {
	Domain   *string `json:"domain,omitempty"`
	Password *string `json:"password,omitempty"`
	Email    *string `json:"email,omitempty"`
}
