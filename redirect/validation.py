import re
import socket

class Validator:
    def __init__(self, params):
        self.params = params
        self.errors = []

    def new_username(self):
        username = self.username()
        if username is not None:
            if not re.match("^[\w-]+$", username):
                self.errors.append("username has invalid characters")
            if len(username) < 5:
                self.errors.append('username is too short (< 5)')
            if len(username) > 50:
                self.errors.append('username is too long (> 50)')
        return username

    def username(self):
        if 'username' in self.params:
            return self.params['username']
        else:
            self.errors.append('missing username')
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
    username = validator.new_username()
    email = validator.email()
    password = validator.new_password()
    port = validator.port()
    ip = validator.ip()
    return validator.errors, username, email, password, port, ip

def update(params):
    validator = Validator(params)
    token = validator.token()
    ip = validator.ip()
    port = validator.port()
    return validator.errors, token, ip, port

def credentials(params):
    validator = Validator(params)
    username = validator.username()
    password = validator.password()
    return validator.errors, username, password

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