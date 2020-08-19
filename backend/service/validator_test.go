package service

import (
	"github.com/stretchr/testify/assert"
	"github.com/syncloud/redirect/model"
	"testing"
)

/*valid_params := {
'user_domain': 'username',
'email': 'valid@mail.com',
'password': 'pass123456',
'port': '80',
'ip': '192.168.1.1'}*/

func UsernameError(t *testing.T, request model.DomainUpdateRequest) *string {
	validator := NewValidator()
	value := validator.newUserDomain(request.UserDomain, true)
	assert.Equal(t, len(validator.errors), 1)
	return value
}

/*
func assertEmailError(t *testing.T, params):

validator = Validator(params)
value = validator.email()
self.assertEqual(len(validator.errors), 1)
return value

func assertNewPasswordError(t *testing.T, params):

validator = Validator(params)
value = validator.new_password()
self.assertEqual(len(validator.errors), 1)
return value
*/

func assertPortError(t *testing.T, request model.DomainUpdateRequest) *int {
	validator := NewValidator()
	value := validator.webPort(request.WebPort)
	assert.Equal(t, 1, len(validator.errors))
	return value
}

/*
func test_new_user_domain_missing(t *testing.T) {

params = {}
user_domain = self.assertUsernameError(params)
self.assertIsNone(user_domain)

func test_new_user_domain_invalid(t *testing.T) {

params = {'user_domain': 'user.name'}
self.assertUsernameError(params)

func test_user_domain_short(t *testing.T) {

params = {'user_domain': 'use'}
self.assertUsernameError(params)

func test_user_domain_long(t *testing.T) {

params = {'user_domain': '12345678901234567890123456789012345678901234567890_'}
self.assertUsernameError(params)

func test_email_missing(t *testing.T) {

params = {}
self.assertEmailError(params)

func test_email_invalid(t *testing.T) {

params = {'email': 'invalid.email'}
self.assertEmailError(params)

func test_password_missing(t *testing.T) {

params = {}
self.assertNewPasswordError(params)

func test_password_short(t *testing.T) {

params = {'password': '123456'}
self.assertNewPasswordError(params)

func test_ip_missing(t *testing.T) {

params = {}
validator = Validator(params)
ip = validator.ip()
self.assertIsNone(ip)
self.assertEquals(0, len(validator.errors))

func test_ip_default(t *testing.T):

params = {}
validator = Validator(params)
ip = validator.ip('192.168.0.1')
self.assertEquals(ip, '192.168.0.1')
self.assertEquals(0, len(validator.errors))

func test_ip_invalid(t *testing.T) {

params = {'ip': '256.256.256.256'}
validator = Validator(params)
ip = validator.ip()
self.assertEqual(len(validator.errors), 1)

func test_port_missing(t *testing.T) {

params = {}
self.assertPortError(params)

func test_port_small(t *testing.T) {

params = {'port': '0'}
self.assertPortError(params)

func test_port_big(t *testing.T) {

params = {'port': '65536'}
self.assertPortError(params)

*/
func testPortNonInt(t *testing.T) {
	port := 1
	assertPortError(t, model.DomainUpdateRequest{WebPort: &port})
}

func TestErrorsAggregated(t *testing.T) {

	request := model.DomainUpdateRequest{}
	validator := NewValidator()
	validator.userDomain(request.UserDomain, true)
	validator.password(request.Password)
	assert.Equal(t, 2, len(validator.errors))
}
