package model

type DomainAvailabilityRequest struct {
	UserDomain *string `json:"user_domain,omitempty"`
	Password   *string `json:"password,omitempty"`
	Email      *string `json:"email,omitempty"`
}
