package validation

import (
	"fmt"
	"github.com/syncloud/redirect/model"
	"log"
	"net"
	"regexp"
	"strings"
)

type FieldValidator struct {
	errors       []string
	fieldsErrors map[string][]string
}

func New() *FieldValidator {
	return &FieldValidator{nil, make(map[string][]string)}
}

func (v *FieldValidator) ToParametersMessages() *[]model.ParameterMessages {
	var messages []model.ParameterMessages
	for k, v := range v.fieldsErrors {
		messages = append(messages, model.ParameterMessages{Parameter: k, Messages: v})
	}
	return &messages
}

func (v *FieldValidator) HasErrors() bool {
	return len(v.fieldsErrors) > 0
}

func (v *FieldValidator) addFieldError(field string, error string) {
	v.errors = append(v.errors, fmt.Sprintf("%s %s", field, error))
	newErrors := []string{error}
	if val, ok := v.fieldsErrors[field]; ok {
		newErrors = append(val, error)
	}
	v.fieldsErrors[field] = newErrors
}

func (v *FieldValidator) Domain(domain *string, field string, mainDomain string) {
	if domain == nil {
		v.addFieldError(field, "Missing")
	} else {
		if *domain == mainDomain {
			v.addFieldError(field, "Invalid domain")
		} else {
			suffix := fmt.Sprintf(".%s", mainDomain)
			if strings.HasSuffix(*domain, suffix) {
				parts := strings.Split(*domain, suffix)
				subDomain := parts[0]
				var valid = regexp.MustCompile(`^[\w-]+$`)
				if !valid.MatchString(subDomain) {
					v.addFieldError(field, "Invalid characters")
				}
				if len(subDomain) < 5 {
					v.addFieldError(field, "Too short (< 5)")
				}
				if len(subDomain) > 50 {
					v.addFieldError(field, "Too long (> 50)")
				}
			}
		}
	}
}

func (v *FieldValidator) Email(email *string) *string {
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

func (v *FieldValidator) NewPassword(newPassword *string) *string {
	password := v.Password(newPassword)
	if password != nil {
		if len(*password) < 7 {
			v.addFieldError("password", "Should be 7 or more characters")
		}
	}
	return password
}

func (v *FieldValidator) Password(password *string) *string {
	if password == nil {
		v.addFieldError("password", "Missing")
	}
	return password
}

func (v *FieldValidator) WebProtocol(webProtocol *string) *string {
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

func (v *FieldValidator) WebLocalPort(webLocalPort *int) *int {
	return v.validatePort(webLocalPort, "web_local_port")
}

func (v *FieldValidator) webPort(webPort *int) *int {
	return v.validatePort(webPort, "web_port")
}

func (v *FieldValidator) validatePort(port *int, field string) *int {
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

func (v *FieldValidator) Token(token *string) {
	if token == nil {
		v.addFieldError("token", "Missing")
	}
}

func (v *FieldValidator) DeviceName(deviceName *string) {
	if deviceName == nil {
		v.addFieldError("device_name", "Missing")
	}
}

func (v *FieldValidator) DeviceTitle(deviceTitle *string) {
	if deviceTitle == nil {
		v.addFieldError("device_title", "Missing")
	}
}

func (v *FieldValidator) checkIpAddress(name string, ip string) {
	if net.ParseIP(ip) == nil {
		v.addFieldError(name, "Invalid IP address")
	}
}

func (v *FieldValidator) Ip(requestIp *string, defaultIp *string) *string {
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

func (v *FieldValidator) LocalIp(localIp *string) {
	if localIp != nil {
		v.checkIpAddress("local_ip", *localIp)
	}
}

func (v *FieldValidator) DeviceMacAddress(deviceMacAddress *string) *string {
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

/*func (v *FieldValidator) string(parameter string, required bool) {
	if val, ok := v.request.:
		if required:
			self.add_field_error(parameter, 'Missing')
		return None
	return self.params[parameter]
}
*/
/*
func (v *FieldValidator) boolean(parameter, required=False, default=None) {
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
