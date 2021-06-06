from datetime import datetime

from models import User, Domain, ActionType
from validation import Validator
import servicesexceptions
import util
import requests


def check_validator(validator):
    if validator.has_errors():
        raise servicesexceptions.parameters_error(validator.fields_errors)


class UsersRead:

    def __init__(self, create_storage, domain):
        self.main_domain = domain
        self.create_storage = create_storage

    def get_user(self, email):
        with self.create_storage() as storage:
            return storage.get_user_by_email(email)

    def authenticate(self, request):
        validator = Validator(request)
        email = validator.email()
        password = validator.password()
        check_validator(validator)

        user = self.get_user(email)
        if not user or not user.active or not util.hash(password) == user.password_hash:
            raise servicesexceptions.bad_request('Authentication failed')

        return user


class Users(UsersRead):

    def __init__(self, create_storage, activate_by_email, mail, dns, domain):
        UsersRead.__init__(self, create_storage, domain)
        self.activate_by_email = activate_by_email
        self.mail = mail
        self.dns = dns
        self.main_domain = domain
        self.create_storage = create_storage

    def get_user(self, email):
        with self.create_storage() as storage:
            return storage.get_user_by_email(email)

    def authenticate(self, request):
        validator = Validator(request)
        email = validator.email()
        password = validator.password()
        check_validator(validator)

        user = self.get_user(email)
        if not user or not user.active or not util.hash(password) == user.password_hash:
            raise servicesexceptions.bad_request('Authentication failed')

        return user

    def user_set_password(self, request):
        validator = Validator(request)
        token = validator.token()
        password = validator.new_password()
        check_validator(validator)

        with self.create_storage() as storage:
            user = storage.get_user_by_token(ActionType.PASSWORD, token)

            if not user:
                raise servicesexceptions.bad_request('Invalid password token')

            user.password_hash = util.hash(password)

            self.mail.send_set_password(user.email)

            action = storage.get_action(token)
            storage.delete(action)
