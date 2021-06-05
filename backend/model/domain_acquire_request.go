package model

import "fmt"

type DomainAcquireRequest struct {
	DeprecatedUserDomain *string `json:"user_domain,omitempty"`
	Domain               *string `json:"domain,omitempty"`
	Password             *string `json:"password,omitempty"`
	Email                *string `json:"email,omitempty"`
	DeviceMacAddress     *string `json:"device_mac_address,omitempty"`
	DeviceName           *string `json:"device_name,omitempty"`
	DeviceTitle          *string `json:"device_title,omitempty"`
}

func (r *DomainAcquireRequest) ForwardCompatibleDomain(mainDomain string) {
	if r.DeprecatedUserDomain != nil {
		domain := fmt.Sprintf("%s.%s", *r.DeprecatedUserDomain, mainDomain)
		r.Domain = &domain
	}
}
