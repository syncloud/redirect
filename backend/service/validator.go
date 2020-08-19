package service

import (
	"fmt"
	"github.com/syncloud/redirect/model"
	"net"
	"regexp"
	"strings"
)

type Validator struct {
	errors       []string
	fieldsErrors map[string][]string
}

func NewValidator() *Validator {
	return &Validator{nil, make(map[string][]string)}
}

func (v *Validator) ToParametersMessages() *[]model.ParameterMessages {
	var messages []model.ParameterMessages
	for k, v := range v.fieldsErrors {
		messages = append(messages, model.ParameterMessages{Parameter: k, Messages: v})
	}
	return &messages
}

func (v *Validator) HasErrors() bool {
	return len(v.fieldsErrors) > 0
}

func (v *Validator) addFieldError(field string, error string) {
	v.errors = append(v.errors, fmt.Sprintf("%s %s", field, error))
	newErrors := []string{error}
	if val, ok := v.fieldsErrors[field]; ok {
		newErrors = append(val, error)
	}
	v.fieldsErrors[field] = newErrors
}

func (v *Validator) newUserDomain(userDomain *string, errorIfMissing bool) *string {
	var valid = regexp.MustCompile(`^[\w-]+$`)
	v.userDomain(userDomain, errorIfMissing)
	if userDomain != nil {
		if !valid.MatchString(*userDomain) {
			v.addFieldError("user_domain", "Invalid characters")
		}
		if len(*userDomain) < 5 {
			v.addFieldError("user_domain", "Too short (< 5)")
		}
		if len(*userDomain) > 50 {
			v.addFieldError("user_domain", "Too long (> 50)")
		}
	}
	return userDomain
}

func (v *Validator) userDomain(userDomain *string, errorIfMissing bool) {
	if userDomain == nil && errorIfMissing {
		v.addFieldError("user_domain", "Missing")
	}
}

/*
func (v *Validator) email(self) {
	if 'email' in
	self.params:
	email = self.params['email']
	if not re.match(r
	"[^@]+@[^@]+\.[^@]+", email):
	self.add_field_error('email', 'Not valid email')
	else:
	return email.lower()
	else:
	self.add_field_error('email', 'Missing')
	return None
}

func (v *Validator) new_password(self) {
	password = self.password()
	if password is
	not
None:
	if len(password) < 7:
	self.add_field_error('password', 'Should be 7 or more characters')
	return password
}
*/

func (v *Validator) password(password *string) {
	if password == nil {
		v.addFieldError("password", "Missing")
	}
}

func (v *Validator) webProtocol(webProtocol *string) *string {
	if webProtocol == nil {
		v.addFieldError("web_protocol", "Missing")
		return nil
	}
	protocol := webProtocol
	protocolLower := strings.ToLower(*protocol)
	if protocolLower != "http" && protocolLower != "https" {
		v.addFieldError("web_protocol", "Protocol should be either http or https")
		return nil
	}
	return &protocolLower
}

func (v *Validator) webLocalPort(webLocalPort *int) *int {
	return v.validatePort(webLocalPort, "web_local_port")
}

func (v *Validator) webPort(webPort *int) *int {
	return v.validatePort(webPort, "web_port")
}

func (v *Validator) validatePort(port *int, field string) *int {
	if port == nil {
		v.addFieldError(field, "Missing")
		return nil
	}

	if *port < 1 || *port > 65535 {
		v.addFieldError(field, "Should be between 1 and 65535")
		return nil
	}
	return port
}

func (v *Validator) Token(token *string) {
	if token == nil {
		v.addFieldError("token", "Missing")
	}
}

func (v *Validator) check_ip_address(name string, ip string) {
	if net.ParseIP(ip) == nil {
		v.addFieldError(name, "Invalid IP address")
	}
}

func (v *Validator) Ip(requestIp *string, defaultIp *string) *string {
	ip := defaultIp
	if requestIp != nil {
		ip = requestIp
	}
	if ip == nil {
		v.addFieldError("ip", "Missing")
		return nil
	}
	v.check_ip_address("ip", *ip)
	return ip
}

func (v *Validator) localIp(localIp *string) {
	if localIp != nil {
		v.check_ip_address("local_ip", *localIp)
	}
}

/*

func (v *Validator) device_mac_address() {
	mac_address = 'device_mac_address'
	if mac_address not
	in
	self.params:
	self.add_field_error(mac_address, 'Missing')
	return None
	mac_address_value = self.params[mac_address]
	if not re.match('[0-9a-f]{2}([-:])[0-9a-f]{2}(\\1[0-9a-f]{2}){4}$', mac_address_value):
	self.add_field_error(mac_address, 'MAC address has wrong format')
	return None
	return mac_address_value
}
*/

/*func (v *Validator) string(parameter string, required bool) {
	if val, ok := v.request.:
		if required:
			self.add_field_error(parameter, 'Missing')
		return None
	return self.params[parameter]
}
*/
/*
func (v *Validator) boolean(parameter, required=False, default=None) {
	if parameter not
	in
	self.params:
	if required:
	self.add_field_error(parameter, 'Missing')
	return default
value = self.params[parameter]
if isinstance(value, basestring):
if value.lower() == 'true':
return True
if value.lower() == 'false':
return False
self.add_field_error(parameter, 'Unexpected value') else:
return value
}

*/
