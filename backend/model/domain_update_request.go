package model

type DomainUpdateRequest struct {
	Ip              *string `json:"ip,omitempty"`
	LocalIp         *string `json:"local_ip,omitempty"`
	MapLocalAddress bool    `json:"map_local_address,omitempty"`
	Token           *string `json:"token,omitempty"`
	Ipv6            *string `json:"ipv6,omitempty"`
	DkimKey         *string `json:"dkim_key,omitempty"`
	PlatformVersion *string `json:"platform_version,omitempty"`
	WebProtocol     *string `json:"web_protocol,omitempty"`
	WebLocalPort    *int    `json:"web_local_port,omitempty"`
	WebPort         *int    `json:"web_port,omitempty"`
	Ipv4Enabled     bool    `json:"ipv4_enabled,omitempty"`
	Ipv6Enabled     bool    `json:"ipv6_enabled,omitempty"`
}
