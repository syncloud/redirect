import MySQLdb

from dbcontext import DbContext
from models import User

def to_user(params):
    (user_domain, email, password_hash, update_token, ip, port, active, activate_token) = params
    return User(user_domain, update_token, ip, port, email, password_hash, active == 1, activate_token)

class UserStorage:

    def __init__(self, mysql_host, mysql_user, mysql_passwd, mysql_db):
        self.mysql_host = mysql_host
        self.mysql_user = mysql_user
        self.mysql_passwd = mysql_passwd
        self.mysql_db = mysql_db

    def connect(self):
        return MySQLdb.connect(self.mysql_host, self.mysql_user, self.mysql_passwd, self.mysql_db)

    select_fields = 'user_domain, email, password_hash, update_token, ip, port, active, activate_token'

    def insert_user(self, user):
        with DbContext(self.connect()) as cursor:
            q = 'insert into user ({0}) values (%s, %s, %s, %s, %s, %s, %s, %s)'.format(self.select_fields)
            p = (user.user_domain, user.email, user.password_hash, user.update_token, user.ip, user.port, user.active, user.activate_token)
            cursor.execute(q, p)

    def update_user(self, user):
        updates = []
        if user.updated_ip_port:
            updates.append('ip = %s' % user.ip)
            updates.append('port = %s' % user.port)

        if user.updated_active:
            updates.append('active = %s' % int(user.active))

        if len(updates) > 0:
            with DbContext(self.connect()) as cursor:
                cursor.execute('update user set ' + ', '.join(updates) + ' where email = %s', user.email)

    def delete_user(self, email):
        with DbContext(self.connect()) as cursor:
            return cursor.execute("delete from user where email = %s", email) > 0

    def get_user_by_email(self, email):
        with DbContext(self.connect()) as cursor:
            num = cursor.execute('select {0} from user where email = %s'.format(self.select_fields), email)
            if num == 1:
                return to_user(cursor.fetchone())
            else:
                return None

    def get_user_by_update_token(self, update_token):
        with DbContext(self.connect()) as cursor:
            num = cursor.execute('select {0} from user where update_token = %s'.format(self.select_fields), update_token)
            if num == 1:
                return to_user(cursor.fetchone())
            else:
                return None

    def get_user_by_domain(self, user_domain):
        with DbContext(self.connect()) as cursor:
            num = cursor.execute('select {0} from user where user_domain = %s'.format(self.select_fields), user_domain)
            if num == 1:
                return to_user(cursor.fetchone())
            else:
                return None

    def get_user_by_activate_token(self, activate_token):
        with DbContext(self.connect()) as cursor:
            num = cursor.execute('select {0} from user where activate_token = %s'.format(self.select_fields), activate_token)
            if num == 1:
                return to_user(cursor.fetchone())
            else:
                return None
