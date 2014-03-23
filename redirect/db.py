"""
DB Querying class
Not using any string concatenation to protect against sql injection attack
"""
import MySQLdb
from dbcontext import DbContext


class Db:

    def __init__(self, mysql_host, mysql_user, mysql_passwd, mysql_db):
        self.mysql_host = mysql_host
        self.mysql_user = mysql_user
        self.mysql_passwd = mysql_passwd
        self.mysql_db = mysql_db

    def connect(self):
        return MySQLdb.connect(self.mysql_host, self.mysql_user, self.mysql_passwd, self.mysql_db)

    def exists(self, username, email):

        with DbContext(self.connect()) as cursor:
            return cursor.execute("""
                select username from user where username = %s or email = %s
                """, (username, email)) > 0

    def insert(self, username, email, password, token, ip, port):

        with DbContext(self.connect()) as cursor:
            cursor.execute("""
                        insert into user (username, email, password_hash, update_token, ip, port, active)
                        values (%s, %s, password(%s), %s, %s, %s, %s)
                    """, (username, email, password, token, ip, port, 0))

    def update(self, token, ip, port):

        with DbContext(self.connect()) as cursor:
            cursor.execute("update user set ip = %s, port = %s where update_token = %s and active = 1", (ip, port, token))

    def existing_token(self, token):

        with DbContext(self.connect()) as cursor:
            return cursor.execute("select update_token from user where update_token = %s and active = 1", token) > 0

    def valid_user(self, username, password):

        with DbContext(self.connect()) as cursor:
            return cursor.execute(
                "select username from user where username = %s and password_hash = password(%s) and active = 1",
                (username, password)) > 0

    def delete_user(self, username, password):

        with DbContext(self.connect()) as cursor:
            return cursor.execute(
                "delete from user where username = %s and password_hash = password(%s) and active = 1", (username, password)) > 0

    def activate(self, token):

        with DbContext(self.connect()) as cursor:
            return cursor.execute("update user set active = 1 where update_token = %s", token) > 0

    def get_port_by_username(self, username):

        with DbContext(self.connect()) as cursor:
            num = cursor.execute('select port from user where username = %s and active = 1', username)
            if num == 1:
                return cursor.fetchone()
            else:
                return None

    def get_user_info_by_token(self, token):

        with DbContext(self.connect()) as cursor:
            num = cursor.execute('select username, ip, port from user where update_token = %s', token)
            if num == 1:
                return cursor.fetchone()
            else:
                return None

    def get_user_info_by_password(self, username, password):

        with DbContext(self.connect()) as cursor:
            num = cursor.execute(
                'select username, ip, port from user where username = %s and password_hash = password(%s) and active=1',
                (username, password))
            if num == 1:
                return cursor.fetchone()
            else:
                return None

    def get_token_by_password(self, username, password):

        with DbContext(self.connect()) as cursor:
            num = cursor.execute(
                'select update_token from user where username = %s and password_hash = password(%s) and active = 1',
                (username, password))
            if num == 1:
                return cursor.fetchone()
            else:
                return None
