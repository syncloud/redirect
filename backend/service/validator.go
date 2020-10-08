package service

import (
	"fmt"
	"github.com/syncloud/redirect/model"
	"log"
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

func (v *Validator) newUserDomain(userDomain *string) *string {
	var valid = regexp.MustCompile(`^[\w-]+$`)
	v.userDomain(userDomain)
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

func (v *Validator) newDomain(domain *string, ourDomain string) *string {
	if domain == nil {
		v.addFieldError("domain", "Missing")
		return nil
	}
	var valid = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{1,61}[a-zA-Z0-9]\.[a-zA-Z]{2,}$`)
	if !valid.MatchString(*domain) {
		v.addFieldError("domain", "Invalid domain name")
	}
	if *domain == ourDomain {
		v.addFieldError("domain", "Cannot use Syncloud domain")
	}
	return domain
}

func (v *Validator) userDomain(userDomain *string) {
	if userDomain == nil {
		v.addFieldError("user_domain", "Missing")
	}
}

func (v *Validator) email(email *string) *string {
	var valid = regexp.MustCompile(`[^@]+@[^@]+\.[^@]+`)
	if email != nil {
		if !valid.MatchString(*email) {
			v.addFieldError("email", "Not valid email")
		} else {
			lower := strings.ToLower(*email)
			return &lower
		}
	} else {
		v.addFieldError("email", "Missing")
	}
	return nil
}

func (v *Validator) newPassword(newPassword *string) *string {
	password := v.password(newPassword)
	if password != nil {
		if len(*password) < 7 {
			v.addFieldError("password", "Should be 7 or more characters")
		}
	}
	return password
}

func (v *Validator) password(password *string) *string {
	if password == nil {
		v.addFieldError("password", "Missing")
	}
	return password
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

func (v *Validator) deviceName(deviceName *string) {
	if deviceName == nil {
		v.addFieldError("device_name", "Missing")
	}
}

func (v *Validator) deviceTitle(deviceTitle *string) {
	if deviceTitle == nil {
		v.addFieldError("device_title", "Missing")
	}
}

func (v *Validator) checkIpAddress(name string, ip string) {
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
	v.checkIpAddress("ip", *ip)
	return ip
}

func (v *Validator) localIp(localIp *string) {
	if localIp != nil {
		v.checkIpAddress("local_ip", *localIp)
	}
}

func (v *Validator) deviceMacAddress(deviceMacAddress *string) *string {
	field := "device_mac_address"
	var pattern = regexp.MustCompile(`^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$`)
	if deviceMacAddress == nil {
		v.addFieldError(field, "Missing")
		return nil
	}
	if !pattern.MatchString(*deviceMacAddress) {
		log.Println("wrong mac", *deviceMacAddress)
		v.addFieldError(field, "MAC address has wrong format")
		return nil
	}
	return deviceMacAddress
}

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
