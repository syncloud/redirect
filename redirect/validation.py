import re
import socket


class Validator:
    def __init__(self):
        pass

    def validate_create(self, params, remote_addr):

        errors = []
        username = None
        if not 'username' in params:
            errors.append('missing username')
        else:
            username = params['username']
            if not re.match("^[\w-]+$", username):
                errors.append("username has invalid characters")
            if len(username) < 5:
                errors.append('username is too short (< 5)')
            if len(username) > 50:
                errors.append('username is too long (> 50)')

        email = None
        if not 'email' in params:
            errors.append('missing email')
        else:
            email = params['email']
            if not re.match(r"[^@]+@[^@]+\.[^@]+", email):
                errors.append('not valid email')

        password = None
        if not 'password' in params:
            errors.append('missing password')
        else:
            password = params['password']
            if len(password) < 7:
                errors.append('password should be 7 or more characters')

        (port_errors, port) = self.validate_port(params)
        errors += port_errors

        (ip_errors, ip) = self.validate_ip(params, remote_addr)
        errors += ip_errors

        return errors, username, email, password, port, ip

    def validate_update(self, params, remote_addr):

        (token_errors, token) = self.validate_token(params)
        (ip_errors, ip) = self.validate_ip(params, remote_addr)
        (port_errors, port) = self.validate_port(params)

        errors = token_errors + ip_errors + port_errors

        return errors, token, ip, port

    def validate_credentials(self, params):

        errors = []
        (username_errors, username) = self.validate_username(params)
        errors += username_errors

        (password_errors, password) = self.validate_password(params)
        errors += password_errors

        return errors, username, password

    def validate_password(self, params):

        errors = []
        password = None
        if not 'password' in params:
            errors.append('missing password')
        else:
            password = params['password']

        return errors, password

    def validate_username(self, params):

        errors = []
        username = None
        if not 'username' in params:
            errors.append('missing username')
        else:
            username = params['username']

        return errors, username

    def validate_port(self, params):
        errors = []
        port = None
        if not 'port' in params:
            errors.append('missing port')
        else:
            port = params['port']
            if not port.isdigit() or int(port) < 1 or int(port) > 65535:
                errors.append('port should a number between 1 and 65535')
        return errors, port

    def validate_ip(self, params, remote_addr):
        errors = []

        if 'ip' in params:
            ip = params['ip']
            try:
                socket.inet_aton(ip)
            except socket.error:
                errors.append('invalid ip')
        else:
            ip = remote_addr
        return errors, ip

    def validate_token(self, params):
        errors = []
        token = None
        if not 'token' in params:
            errors.append('No token provided')
        else:
            token = params['token']

        return errors, token