import re
import socket


class Validator:
    def __init__(self, params):
        self.params = params
        self.errors = []
        self.fields_errors = {}

    def has_errors(self):
        return len(self.fields_errors) > 0

    def add_field_error(self, field, error):
        self.errors.append("{0} {1}".format(field, error))
        if field not in self.fields_errors:
            self.fields_errors[field] = []
        self.fields_errors[field].append(error)

    def new_user_domain(self, error_if_missing=True):
        user_domain = self.user_domain(error_if_missing)
        if user_domain is not None:
            if not re.match("^[\w-]+$", user_domain):
                self.add_field_error("user_domain", "Invalid characters")
            if len(user_domain) < 5:
                self.add_field_error("user_domain", "Too short (< 5)")
            if len(user_domain) > 50:
                self.add_field_error("user_domain", "Too long (> 50)")
        return user_domain

    def user_domain(self, error_if_missing=True):
        if 'user_domain' in self.params:
            return self.params['user_domain']
        else:
            if error_if_missing:
                self.add_field_error("user_domain", "Missing")
        return None

    def email(self):
        if 'email' in self.params:
            email = self.params['email']
            if not re.match(r"[^@]+@[^@]+\.[^@]+", email):
                self.add_field_error('email', 'Not valid email')
            else:
                return email
        else:
            self.add_field_error('email', 'Missing')
        return None

    def new_password(self):
        password = self.password()
        if password is not None:
            if len(password) < 7:
                self.add_field_error('password', 'Should be 7 or more characters')
        return password

    def password(self):
        if 'password' not in self.params:
            self.add_field_error('password', 'Missing')
            return None
        return self.params['password']

    def port(self):
        if 'port' not in self.params:
            self.add_field_error('port', 'Missing')
            return None
        try:
            port_num = int(self.params['port'])
        except:
            self.add_field_error('port', 'Should be a number')
            return None
        if port_num < 1 or port_num > 65535:
            self.add_field_error('port', 'Should be between 1 and 65535')
            return None
        return port_num

    def token(self, token='token'):
        if token not in self.params:
            self.add_field_error(token, 'Missing')
            return None
        return self.params[token]

    def __check_ip_address(self, name, ip):
        try:
            socket.inet_aton(ip)
        except socket.error:
            self.add_field_error(name, 'Invalid IP address')

    def ip(self, default_ip=None):
        ip = default_ip
        if 'ip' in self.params:
            ip = self.params['ip']
        if not ip:
            return None

        self.__check_ip_address('ip', ip)

        return ip

    def local_ip(self, default_ip=None):
        ip = default_ip
        if 'local_ip' in self.params:
            ip = self.params['local_ip']
        if not ip:
            return None

        self.__check_ip_address('local_ip', ip)

        return ip

    def device_mac_address(self):
        mac_address = 'device_mac_address'
        if mac_address not in self.params:
            self.add_field_error(mac_address, 'Missing')
            return None
        return self.params[mac_address]

    def string(self, parameter):
        if parameter not in self.params:
            return None
        return self.params[parameter]
