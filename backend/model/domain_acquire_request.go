package model

type DomainAcquireRequest struct {
	UserDomain       *string `json:"user_domain,omitempty"`
	Password         *string `json:"password,omitempty"`
	Email            *string `json:"email,omitempty"`
	DeviceMacAddress *string `json:"device_mac_address,omitempty"`
	DeviceName       *string `json:"device_name,omitempty"`
	DeviceTitle      *string `json:"device_title,omitempty"`
}
