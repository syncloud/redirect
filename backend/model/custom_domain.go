package model

import (
	"time"
)

type (
	CustomDomain struct {
		Id           uint64     `json:"-"`
		Domain       string     `json:"domain,omitempty"`
		Ip           *string    `json:"ip,omitempty"`
		Ipv6         *string    `json:"ipv6,omitempty"`
		DkimKey      *string    `json:"dkim_key,omitempty"`
		UpdateToken  *string    `json:"update_token,omitempty"`
		LastUpdate   *time.Time `json:"last_update,omitempty"`
		Port         *int       `json:"port,omitempty"`
		UserId       uint64     `json:"-"`
		HostedZoneId *string    `json:"-"`
	}
)
