import MySQLdb
from dbcontext import DbContext
import hashlib

def hash(plain_password):
    return hashlib.sha256(plain_password).hexdigest()

class User:

    def __init__(self, user_domain, email, password_hash, update_token, ip, port, active):
        self.user_domain = user_domain
        self.email = email
        self.password_hash = password_hash
        self.update_token = update_token
        self.ip = ip
        self.port = port
        self.active = active

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

def to_user(params):
    (user_domain, email, password_hash, update_token, ip, port, active) = params
    return User(user_domain, email, password_hash, update_token, ip, port, active == 1)

class UserStorage:

    def __init__(self, mysql_host, mysql_user, mysql_passwd, mysql_db):
        self.mysql_host = mysql_host
        self.mysql_user = mysql_user
        self.mysql_passwd = mysql_passwd
        self.mysql_db = mysql_db

    def connect(self):
        return MySQLdb.connect(self.mysql_host, self.mysql_user, self.mysql_passwd, self.mysql_db)

    def insert_user(self, user):
        with DbContext(self.connect()) as cursor:
            cursor.execute("""
                        insert into user (user_domain, email, password_hash, update_token, ip, port, active)
                        values (%s, %s, %s, %s, %s, %s, %s)
                    """, (user.user_domain, user.email, user.password_hash, user.update_token, user.ip, user.port, user.active))

    def update_user(self, user):
        updates = []
        if user.updated_ip_port:
            updates.append('ip = %s' % user.ip)
            updates.append('port = %s' % user.port)

        if user.updated_active:
            updates.append('active = %s' % int(user.active))

        if len(updates) > 0:
            query = 'update user set ' + ', '.join(updates) + (' where email = %s' % user.email)
            with DbContext(self.connect()) as cursor:
                cursor.execute(query)

    def delete_user(self, email):
        with DbContext(self.connect()) as cursor:
            return cursor.execute("delete from user where email = %s", email) > 0

    def get_user_by_email(self, email):
        with DbContext(self.connect()) as cursor:
            num = cursor.execute('select user_domain, email, password_hash, update_token, ip, port, active from user where email = %s', email)
            if num == 1:
                return to_user(cursor.fetchone())
            else:
                return None

    def get_user_by_token(self, update_token):
        with DbContext(self.connect()) as cursor:
            num = cursor.execute('select user_domain, email, password_hash, update_token, ip, port, active from user where update_token = %s', update_token)
            if num == 1:
                return to_user(cursor.fetchone())
            else:
                return None