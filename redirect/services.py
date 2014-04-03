import uuid
from models import User
from validation import Validator
from util import hash
import servicesexceptions

class Users:
    def __init__(self, user_storage):
        self.storage = user_storage

    def create_new_user(self, request):
        validator = Validator(request)
        email = validator.email()
        password = validator.password()
        user_domain = validator.new_user_domain()
        errors = validator.errors

        if errors:
            message = ", ".join(errors)
            raise servicesexceptions.bad_request(message)

        by_email = self.storage.get_user_by_email(email)
        if by_email and by_email.email == email:
            raise servicesexceptions.conflict('Email is already registered')

        by_domain = self.storage.get_user_by_domain(user_domain)
        if by_domain and by_domain.user_domain == user_domain:
            raise servicesexceptions.conflict('User domain name is already in use')

        update_token = uuid.uuid4().hex
        activate_token = uuid.uuid4().hex
        user = User(user_domain, update_token, None, None, email, hash(password), False, activate_token)

        self.storage.insert_user(user)

        return user

    def activate(self, request):
        validator = Validator(request)
        token = validator.token()
        errors = validator.errors

        user = self.storage.get_user_by_activate_token(token)
        if not user:
            raise servicesexceptions.bad_request('Invalid activation token')

        if errors:
            message = ", ".join(errors)
            raise servicesexceptions.bad_request(message)

        user.update_active(True)
        self.storage.update_user(user)

        return True
