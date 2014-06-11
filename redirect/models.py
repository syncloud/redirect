from storm.locals import Bool, Int, Unicode, RawStr

class User(object):
    __storm_table__ = "user"
    email = Unicode(primary=True)
    password_hash = RawStr()
    active = Bool()
    activate_token = Unicode()
    user_domain = Unicode()
    ip = Unicode()
    update_token = Unicode()
    port = Int()

    def __init__(self, user_domain, update_token, ip, port, email, password_hash, active, activate_token):
        self.user_domain = user_domain
        self.update_token = update_token
        self.ip = ip
        self.port = port
        self.email = email
        self.password_hash = password_hash
        self.active = active
        self.activate_token = activate_token

    def update_ip_port(self, ip, port):
        self.ip = ip
        self.port = port

    def update_active(self, active):
        self.active = active