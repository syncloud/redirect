package service

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNotChanged(t *testing.T) {
	ip := "1"
	ipv6 := "1"
	dkim := "1"
	localIp := "1"

	newIp := "1"
	newIpv6 := "1"
	newDkim := "1"
	newLocalIp := "1"

	changed := Changed(
		true, &ip, &ipv6, &dkim, &localIp,
		true, &newIp, &newIpv6, &newDkim, &newLocalIp)

	assert.False(t, changed)
}

func TestIpChanged(t *testing.T) {
	ip := "1"
	ipv6 := "1"
	dkim := "1"
	localIp := "1"

	newIp := "2"
	newIpv6 := "1"
	newDkim := "1"
	newLocalIp := "1"

	changed := Changed(
		true, &ip, &ipv6, &dkim, &localIp,
		true, &newIp, &newIpv6, &newDkim, &newLocalIp)

	assert.True(t, changed)
}

func TestIpv6Changed(t *testing.T) {
	ip := "1"
	ipv6 := "1"
	dkim := "1"
	localIp := "1"

	newIp := "1"
	newIpv6 := "2"
	newDkim := "1"
	newLocalIp := "1"

	changed := Changed(
		true, &ip, &ipv6, &dkim, &localIp,
		true, &newIp, &newIpv6, &newDkim, &newLocalIp)

	assert.True(t, changed)
}

func TestDkimChanged(t *testing.T) {
	ip := "1"
	ipv6 := "1"
	dkim := "1"
	localIp := "1"

	newIp := "1"
	newIpv6 := "1"
	newDkim := "2"
	newLocalIp := "1"

	changed := Changed(
		true, &ip, &ipv6, &dkim, &localIp,
		true, &newIp, &newIpv6, &newDkim, &newLocalIp)

	assert.True(t, changed)
}

func TestLocalIpChanged(t *testing.T) {
	ip := "1"
	ipv6 := "1"
	dkim := "1"
	localIp := "1"

	newIp := "1"
	newIpv6 := "1"
	newDkim := "1"
	newLocalIp := "2"

	changed := Changed(
		true, &ip, &ipv6, &dkim, &localIp,
		true, &newIp, &newIpv6, &newDkim, &newLocalIp)

	assert.True(t, changed)
}

func TestMapLocalAddressChanged(t *testing.T) {
	ip := "1"
	ipv6 := "1"
	dkim := "1"
	localIp := "1"

	newIp := "1"
	newIpv6 := "1"
	newDkim := "1"
	newLocalIp := "1"

	changed := Changed(
		true, &ip, &ipv6, &dkim, &localIp,
		false, &newIp, &newIpv6, &newDkim, &newLocalIp)

	assert.True(t, changed)
}

func TestEquals(t *testing.T) {
	ip := "127.0.0.1"
	ip1 := "127.0.0.1"

	assert.True(t, Equals(&ip, &ip1))
	assert.True(t, Equals(nil, nil))
	assert.False(t, Equals(&ip, nil))
	assert.False(t, Equals(nil, &ip))
}
