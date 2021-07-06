package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBackwardCompatibleDomain_Free(t *testing.T) {
	domain := &Domain{Name: "test123.syncloud.it"}
	domain.BackwardCompatibleDomain("syncloud.it")
	assert.Equal(t, "test123", domain.DeprecatedUserDomain)
}

func TestBackwardCompatibleDomain_FreeContains(t *testing.T) {
	domain := &Domain{Name: "test123syncloud.it"}
	domain.BackwardCompatibleDomain("syncloud.it")
	assert.Equal(t, "", domain.DeprecatedUserDomain)
}

func TestBackwardCompatibleDomain_Managed(t *testing.T) {
	domain := &Domain{Name: "test123.com"}
	domain.BackwardCompatibleDomain("syncloud.it")
	assert.Equal(t, "", domain.DeprecatedUserDomain)
}

func TestIpv6(t *testing.T) {
	ipv6 := "2001:0db8:85a3:0000:0000:8a2e:0370:7334"
	domain := &Domain{Name: "domain", Ipv6: &ipv6}
	assert.Equal(t, ipv6, *domain.DnsIpv6())
	assert.Nil(t, domain.DnsIpv4())
}

func TestIpv4(t *testing.T) {
	ipv4 := "192.168.0.1"
	domain := Domain{Ip: &ipv4}
	assert.Nil(t, domain.DnsIpv6())
	assert.Equal(t, ipv4, *domain.DnsIpv4())
}

func TestAccessIpExternal(t *testing.T) {
	ipv4 := "192.168.0.1"
	localIp := "192.168.0.2"
	domain := Domain{Ip: &ipv4, LocalIp: &localIp, MapLocalAddress: false}
	assert.Equal(t, ipv4, *domain.accessIp())
}

func TestAccessIpLocal(t *testing.T) {
	ipv4 := "192.168.0.1"
	localIp := "192.168.0.2"
	domain := Domain{Ip: &ipv4, LocalIp: &localIp, MapLocalAddress: true}
	assert.Equal(t, localIp, *domain.accessIp())
}
