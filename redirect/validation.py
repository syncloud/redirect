import re
import socket


class Validator:
    def __init__(self, params):
        self.params = params
        self.errors = []
        self.fields_errors = {}

    def add_field_error(self, field, error):
        self.errors.append("{0} {1}".format(field, error))
        if field not in self.fields_errors:
            self.fields_errors[field] = []
        self.fields_errors[field].append(error)

    def new_user_domain(self):
        user_domain = self.user_domain()
        if user_domain is not None:
            if not re.match("^[\w-]+$", user_domain):
                self.add_field_error("user_domain", "invalid characters")
            if len(user_domain) < 5:
                self.add_field_error("user_domain", "too short (< 5)")
            if len(user_domain) > 50:
                self.add_field_error("user_domain", "too long (> 50)")
        return user_domain

    def user_domain(self):
        if 'user_domain' in self.params:
            return self.params['user_domain']
        else:
            self.add_field_error("user_domain", "missing")
        return None

    def email(self):
        if 'email' in self.params:
            email = self.params['email']
            if not re.match(r"[^@]+@[^@]+\.[^@]+", email):
                self.add_field_error('email', 'not valid email')
            else:
                return email
        else:
            self.add_field_error('email', 'missing email')
        return None

    def new_password(self):
        password = self.password()
        if password is not None:
            if len(password) < 7:
                self.add_field_error('password', 'should be 7 or more characters')
        return password

    def password(self):
        if 'password' not in self.params:
            self.add_field_error('password', 'missing')
            return None
        return self.params['password']

    def port(self):
        if 'port' not in self.params:
            self.add_field_error('port', 'missing')
            return None
        try:
            port_num = int(self.params['port'])
        except:
            self.add_field_error('port', 'should be a number')
            return None
        if port_num < 1 or port_num > 65535:
            self.add_field_error('port', 'should be between 1 and 65535')
            return None
        return port_num

    def token(self):
        if 'token' not in self.params:
            self.add_field_error('token', 'missing')
            return None
        return self.params['token']

    def ip(self, default_ip=None):

        ip = default_ip
        if 'ip' in self.params:
            ip = self.params['ip']
        if not ip:
            return None

        try:
            socket.inet_aton(ip)
        except socket.error:
            self.add_field_error('ip', 'invalid ip')
        return ip