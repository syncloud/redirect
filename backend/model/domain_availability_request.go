package model

type DomainAvailabilityRequest struct {
	Domain   *string `json:"domain,omitempty"`
	Password *string `json:"password,omitempty"`
	Email    *string `json:"email,omitempty"`
}

func (r *DomainAvailabilityRequest) IsFree(mainDomain string) bool {
	return IsFree(*r.Domain, mainDomain)
}
