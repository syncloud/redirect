from models import User
from storm.locals import create_database, Store

class UserStorage:

    def __init__(self, mysql_host, mysql_user, mysql_passwd, mysql_db):
        self.database_spec = "mysql://{0}:{1}@{2}:{3}/{4}".format(mysql_user, mysql_passwd, mysql_host, 3306, mysql_db)
        self.database = create_database(self.database_spec)
        self.store = Store(self.database)

    def insert_user(self, user):
        self.store.add(user)

    def delete_user(self, email):
        user = self.store.find(User, User.email == email).one()
        if user is not None:
            self.store.remove(user)
            return True
        return False

    def get_user_by_email(self, email):
        user = self.store.find(User, User.email == email).one()
        return user

    def get_user_by_update_token(self, update_token):
        user = self.store.find(User, User.update_token == update_token).one()
        return user

    def get_user_by_domain(self, user_domain):
        user = self.store.find(User, User.user_domain == user_domain).one()
        return user

    def get_user_by_activate_token(self, activate_token):
        user = self.store.find(User, User.activate_token == activate_token).one()
        return user

    def save(self):
        self.store.commit()