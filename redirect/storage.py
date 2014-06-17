from models import User, Domain, Service, Base

class Storage:

    def __init__(self, session):
        self.session = session

    def delete_user(self, email):
        user = self.get_user_by_email(email)
        if user is not None:
            self.session.delete(user)
            return True
        return False

    def get_user_by_email(self, email):
        user = self.session.query(User).filter(User.email == email).first()
        return user

    def get_user_by_activate_token(self, activate_token):
        user = self.session.query(User).filter(User.activate_token == activate_token).first()
        return user

    def get_domain_by_update_token(self, update_token):
        domain = self.session.query(Domain).filter(Domain.update_token == update_token).first()
        return domain

    def get_domain_by_name(self, user_domain):
        domain = self.session.query(Domain).filter(Domain.user_domain == user_domain).first()
        return domain

    def add(self, obj):
        self.session.add(obj)

    def clear(self):
        self.session.query(Service).delete()
        self.session.query(Domain).delete()
        self.session.query(User).delete()