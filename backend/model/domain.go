package model

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

type (
	Domain struct {
		Id                   uint64     `json:"-"`
		DeprecatedUserDomain string     `json:"user_domain,omitempty"`
		Ip                   *string    `json:"ip,omitempty"`
		Ipv6                 *string    `json:"ipv6,omitempty"`
		DkimKey              *string    `json:"dkim_key,omitempty"`
		LocalIp              *string    `json:"local_ip,omitempty"`
		MapLocalAddress      bool       `json:"map_local_address"`
		UpdateToken          *string    `json:"update_token,omitempty"`
		LastUpdate           *time.Time `json:"last_update,omitempty"`
		DeviceMacAddress     *string    `json:"device_mac_address,omitempty"`
		DeviceName           *string    `json:"device_name,omitempty"`
		DeviceTitle          *string    `json:"device_title,omitempty"`
		PlatformVersion      *string    `json:"platform_version,omitempty"`
		WebProtocol          *string    `json:"web_protocol,omitempty"`
		WebPort              *int       `json:"web_port,omitempty"`
		WebLocalPort         *int       `json:"web_local_port,omitempty"`
		UserId               int64      `json:"-"`
		HostedZoneId         string     `json:"-"`
		Name                 string     `json:"name,omitempty"`
		NameServers          []*string  `json:"name_servers,omitempty"`
		Timestamp            *time.Time `json:"timestamp,omitempty"`
	}
)

func (d *Domain) BackwardCompatibleDomain(mainDomain string) {
	suffix := fmt.Sprintf(".%s", mainDomain)
	if strings.HasSuffix(d.Name, suffix) {
		parts := strings.Split(d.Name, suffix)
		d.DeprecatedUserDomain = parts[0]
	}
}

func (d *Domain) FQDN() string {
	return fmt.Sprintf("%s.", d.Name)
}

func (d *Domain) accessIp() *string {
	if d.MapLocalAddress {
		return d.LocalIp
	}
	return d.Ip
}

func (d *Domain) DnsIpv6() *string {

	if d.Ipv6 != nil {
		ip := net.ParseIP(*d.Ipv6)
		if ip.To4() == nil && ip.To16() != nil {
			log.Printf("ipv6: %s\n", *d.Ipv6)
			return d.Ipv6
		}
	}
	accessIp := d.accessIp()
	if accessIp != nil {
		ip := net.ParseIP(*accessIp)
		if ip.To4() == nil && ip.To16() != nil {
			log.Printf("ipv6: %s\n", *accessIp)
			return accessIp
		}
	}
	return nil
}

func (d *Domain) DnsIpv4() *string {
	accessIp := d.accessIp()
	if accessIp != nil && net.ParseIP(*accessIp).To4() != nil {
		return accessIp
	}
	return nil
}

func IsFree(domain string, mainDomain string) bool {
	return strings.HasSuffix(domain, fmt.Sprintf(".%s", mainDomain))
}
