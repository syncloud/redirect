package service

import (
	"github.com/stretchr/testify/assert"
	"github.com/syncloud/redirect/model"
	"testing"
)

func TestEmailMissing(t *testing.T) {
	validator := NewValidator()
	_ = validator.email(nil)
	assert.Equal(t, len(validator.errors), 1)
}

func TestEmailInvalid(t *testing.T) {
	validator := NewValidator()
	email := "invalid.email"
	_ = validator.email(&email)
	assert.Equal(t, len(validator.errors), 1)
}

func TestNewUserDomainMissing(t *testing.T) {

	request := model.DomainAcquireRequest{}

	validator := NewValidator()
	result := validator.newUserDomain(request.UserDomain)
	assert.Equal(t, len(validator.errors), 1)
	assert.Nil(t, result)
}

func TestNewUserDomainInvalid(t *testing.T) {
	domain := "user.name"
	request := model.DomainAcquireRequest{UserDomain: &domain}
	validator := NewValidator()
	_ = validator.newUserDomain(request.UserDomain)
	assert.Equal(t, len(validator.errors), 1)
}

func TestUserDomainShort(t *testing.T) {
	domain := "use"
	request := model.DomainAcquireRequest{UserDomain: &domain}
	validator := NewValidator()
	_ = validator.newUserDomain(request.UserDomain)
	assert.Equal(t, len(validator.errors), 1)
}

func TestUserDomainLong(t *testing.T) {

	domain := "12345678901234567890123456789012345678901234567890_"
	request := model.DomainAcquireRequest{UserDomain: &domain}
	validator := NewValidator()
	_ = validator.newUserDomain(request.UserDomain)
	assert.Equal(t, len(validator.errors), 1)
}

func TestPasswordMissing(t *testing.T) {
	validator := NewValidator()
	result := validator.newPassword(nil)
	assert.Equal(t, 1, len(validator.errors))
	assert.Nil(t, result)
}

func TestPasswordShort(t *testing.T) {
	validator := NewValidator()
	password := "123456"
	result := validator.newPassword(&password)
	assert.Equal(t, 1, len(validator.errors))
	assert.Equal(t, "123456", *result)
}

func TestIpMissing(t *testing.T) {
	validator := NewValidator()
	result := validator.Ip(nil, nil)
	assert.Equal(t, 1, len(validator.errors))
	assert.Nil(t, result)
}

func TestIpDefault(t *testing.T) {
	defaultIp := "192.168.0.2"
	validator := NewValidator()
	result := validator.Ip(nil, &defaultIp)
	assert.Equal(t, 0, len(validator.errors))
	assert.Equal(t, *result, "192.168.0.2")
}

func TestIpInvalid(t *testing.T) {
	ip := "256.256.256.256"
	validator := NewValidator()
	_ = validator.Ip(&ip, nil)
	assert.Equal(t, 1, len(validator.errors))
}

func TestPortMissing(t *testing.T) {
	request := model.DomainUpdateRequest{}
	validator := NewValidator()
	_ = validator.webPort(request.WebPort)
	assert.Equal(t, 1, len(validator.errors))
}

func TestPortTooSmall(t *testing.T) {
	port := 0
	request := model.DomainUpdateRequest{WebPort: &port}
	validator := NewValidator()
	_ = validator.webPort(request.WebPort)
	assert.Equal(t, 1, len(validator.errors))
}

func TestPortTooBig(t *testing.T) {
	port := 65536
	request := model.DomainUpdateRequest{WebPort: &port}
	validator := NewValidator()
	_ = validator.webPort(request.WebPort)
	assert.Equal(t, 1, len(validator.errors))
}

func TestErrorsAggregated(t *testing.T) {

	validator := NewValidator()
	validator.userDomain(nil)
	validator.password(nil)
	assert.Equal(t, 2, len(validator.errors))
}

func TestWrongMacAddress(t *testing.T) {

	validator := NewValidator()
	mac := "wrong_mac"
	validator.deviceMacAddress(&mac)
	assert.Equal(t, 1, len(validator.errors))
}

func TestGoodMacAddress(t *testing.T) {

	validator := NewValidator()
	mac := "11:22:33:44:55:66"
	validator.deviceMacAddress(&mac)
	assert.Equal(t, 0, len(validator.errors))
}
