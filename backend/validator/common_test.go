package validator

import (
	"github.com/stretchr/testify/assert"
	"github.com/syncloud/redirect/model"
	"testing"
)

func TestEmailMissing(t *testing.T) {
	validator := New()
	_ = validator.Email(nil)
	assert.Equal(t, len(validator.errors), 1)
}

func TestEmailInvalid(t *testing.T) {
	validator := New()
	email := "invalid.email"
	_ = validator.Email(&email)
	assert.Equal(t, len(validator.errors), 1)
}

func TestDomainMissing(t *testing.T) {

	validator := New()
	result := validator.Domain(nil, "", "syncloud.it")
	assert.Equal(t, len(validator.errors), 1)
	assert.Nil(t, result)
}

func TestFreeDomainInvalid(t *testing.T) {
	domain := "user.name.syncloud.it"
	validator := New()
	_ = validator.Domain(&domain, "", "syncloud.it")
	assert.Equal(t, len(validator.errors), 1)
}

func TestFreeDomainShort(t *testing.T) {
	domain := "use.syncloud.it"
	validator := New()
	_ = validator.Domain(&domain, "", "syncloud.it")
	assert.Equal(t, len(validator.errors), 1)
}

func TestFreeDomainLong(t *testing.T) {
	domain := "12345678901234567890123456789012345678901234567890_.syncloud.it"
	validator := New()
	_ = validator.Domain(&domain, "", "syncloud.it")
	assert.Equal(t, 1, len(validator.errors))
}

func TestFreeDomainEmpty(t *testing.T) {
	domain := ".syncloud.it"
	validator := New()
	_ = validator.Domain(&domain, "", "syncloud.it")
	assert.Equal(t, 2, len(validator.errors))
}

func TestFreeDomainContainsSubdomain(t *testing.T) {
	domain := "test123.test123.syncloud.it"
	validator := New()
	_ = validator.Domain(&domain, "", "syncloud.it")
	assert.Equal(t, 1, len(validator.errors))
}

func TestPremiumEqualsFreeDomain(t *testing.T) {
	domain := "syncloud.it"
	validator := New()
	_ = validator.Domain(&domain, "", "syncloud.it")
	assert.Equal(t, 1, len(validator.errors))
}

func TestPremiumOk(t *testing.T) {
	domain := "example.com"
	validator := New()
	_ = validator.Domain(&domain, "", "syncloud.it")
	assert.Equal(t, 0, len(validator.errors))
}

func TestPasswordMissing(t *testing.T) {
	validator := New()
	result := validator.NewPassword(nil)
	assert.Equal(t, 1, len(validator.errors))
	assert.Nil(t, result)
}

func TestPasswordShort(t *testing.T) {
	validator := New()
	password := "123456"
	result := validator.NewPassword(&password)
	assert.Equal(t, 1, len(validator.errors))
	assert.Equal(t, "123456", *result)
}

func TestIpMissing(t *testing.T) {
	validator := New()
	result := validator.Ip(nil, nil)
	assert.Equal(t, 1, len(validator.errors))
	assert.Nil(t, result)
}

func TestIpDefault(t *testing.T) {
	defaultIp := "192.168.0.2"
	validator := New()
	result := validator.Ip(nil, &defaultIp)
	assert.Equal(t, 0, len(validator.errors))
	assert.Equal(t, *result, "192.168.0.2")
}

func TestIpInvalid(t *testing.T) {
	ip := "256.256.256.256"
	validator := New()
	_ = validator.Ip(&ip, nil)
	assert.Equal(t, 1, len(validator.errors))
}

func TestPortMissing(t *testing.T) {
	request := model.DomainUpdateRequest{}
	validator := New()
	_ = validator.webPort(request.WebPort)
	assert.Equal(t, 1, len(validator.errors))
}

func TestPortTooSmall(t *testing.T) {
	port := 0
	request := model.DomainUpdateRequest{WebPort: &port}
	validator := New()
	_ = validator.webPort(request.WebPort)
	assert.Equal(t, 1, len(validator.errors))
}

func TestPortTooBig(t *testing.T) {
	port := 65536
	request := model.DomainUpdateRequest{WebPort: &port}
	validator := New()
	_ = validator.webPort(request.WebPort)
	assert.Equal(t, 1, len(validator.errors))
}

func TestErrorsAggregated(t *testing.T) {
	validator := New()
	validator.Domain(nil, "", "syncloud.it")
	validator.Password(nil)
	assert.Equal(t, 2, len(validator.errors))
}

func TestWrongMacAddress(t *testing.T) {

	validator := New()
	mac := "wrong_mac"
	validator.DeviceMacAddress(&mac)
	assert.Equal(t, 1, len(validator.errors))
}

func TestGoodMacAddress(t *testing.T) {

	validator := New()
	mac := "11:22:33:44:55:66"
	validator.DeviceMacAddress(&mac)
	assert.Equal(t, 0, len(validator.errors))
}
