from util import hash

class User:

    def __init__(self, user_domain, update_token, ip, port, email, password_hash, active, activate_token):
        self.user_domain = user_domain
        self.email = email
        self.password_hash = password_hash
        self.update_token = update_token
        self.ip = ip
        self.port = port
        self.active = active
        self.activate_token = activate_token

        self.update_ip_port = False
        self.updated_active = False

    def check_active(self, password_plain):
        p = hash(password_plain)
        return self.active and self.password_hash == p

    def update_ip_port(self, ip, port):
        self.ip = ip
        self.port = port
        self.updated_ip_port = True

    def update_active(self, active):
        self.active = active
        self.updated_active = True