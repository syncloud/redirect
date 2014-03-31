import re
import socket

class Validator:
    def __init__(self, params):
        self.params = params
        self.errors = []

    def new_user_domain(self):
        user_domain = self.user_domain()
        if user_domain is not None:
            if not re.match("^[\w-]+$", user_domain):
                self.errors.append("user domain has invalid characters")
            if len(user_domain) < 5:
                self.errors.append('user domain is too short (< 5)')
            if len(user_domain) > 50:
                self.errors.append('user domain is too long (> 50)')
        return user_domain

    def user_domain(self):
        if 'user_domain' in self.params:
            return self.params['user_domain']
        else:
            self.errors.append('missing user domain')
        return None

    def email(self):
        if 'email' in self.params:
            email = self.params['email']
            if not re.match(r"[^@]+@[^@]+\.[^@]+", email):
                self.errors.append('not valid email')
            else:
                return email
        else:
            self.errors.append('missing email')
        return None

    def new_password(self):
        password = self.password()
        if password is not None:
            if len(password) < 7:
                self.errors.append('password should be 7 or more characters')
        return password

    def password(self):
        if 'password' not in self.params:
            self.errors.append('missing password')
            return None
        return self.params['password']

    def port(self):
        if 'port' not in self.params:
            self.errors.append('missing port')
            return None
        port = self.params['port']
        try:
            port_num = int(port)
            if port_num < 1 or port_num > 65535:
                self.errors.append('port should a number between 1 and 65535')
                return None
            return port_num
        except:
            return None

    def token(self):
        if 'token' not in self.params:
            self.errors.append('No token provided')
            return None
        return self.params['token']

    def ip(self):
        if 'ip' not in self.params:
            return None
        ip = self.params['ip']
        try:
            socket.inet_aton(ip)
        except socket.error:
            self.errors.append('invalid ip')
        return ip


def create(params):
    validator = Validator(params)
    user_domain = validator.new_user_domain()
    email = validator.email()
    password = validator.new_password()
    port = validator.port()
    ip = validator.ip()
    return validator.errors, user_domain, email, password, port, ip

def update(params):
    validator = Validator(params)
    token = validator.token()
    ip = validator.ip()
    port = validator.port()
    return validator.errors, token, ip, port

def credentials(params):
    validator = Validator(params)
    user_domain = validator.user_domain()
    password = validator.password()
    return validator.errors, user_domain, password

def token(params):
    validator = Validator(params)
    token = validator.token()
    return validator.errors, token



class Validation:
    def __init__(self):
        pass

    def validate_create(self, params):
        return create(params)

    def validate_update(self, params):
        return update(params)

    def validate_credentials(self, params):
        return credentials(params)

    def validate_token(self, params):
        return token(params)