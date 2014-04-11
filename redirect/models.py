from util import hash

class User:

    def __init__(self, user_domain, update_token, ip, port, email, password_hash, active, activate_token):
        self.user_domain = user_domain
        self.update_token = update_token
        self.ip = ip
        self.port = port
        self.email = email
        self.password_hash = password_hash
        self.active = active
        self.activate_token = activate_token

        self.updated_ip_port = False
        self.updated_active = False

    def update_ip_port(self, ip, port):
        self.ip = ip
        self.port = port
        self.updated_ip_port = True

    def update_active(self, active):
        self.active = active
        self.updated_active = True