package model

import (
	"fmt"
	"strings"
)

type DomainAvailabilityRequest struct {
	Domain   *string `json:"domain,omitempty"`
	Password *string `json:"password,omitempty"`
	Email    *string `json:"email,omitempty"`
}

func (r *DomainAvailabilityRequest) IsFree(mainDomain string) bool {
	return strings.HasSuffix(*r.Domain, fmt.Sprintf(".%s", mainDomain))
}
