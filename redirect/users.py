import uuid
from storage import User
from validation import Validator
from redirectutil import hash
import restexceptions

class Users:
    def __init__(self, user_storage):
        self.storage = user_storage

    def create_new_user(self, request):
        validator = Validator(request)
        email = validator.email()
        password = validator.password()
        user_domain = validator.new_user_domain()
        update_token = uuid.uuid4().hex
        activate_token = uuid.uuid4().hex

        by_email = self.storage.get_user_by_email(email)
        if by_email and by_email.email == email:
            raise restexceptions.conflict('Email is already registered')

        by_domain = self.storage.get_user_by_domain(user_domain)
        if by_domain and by_domain.user_domain == user_domain:
            raise restexceptions.conflict('User domain name is already in use')

        user = User(user_domain, update_token, None, None, email, hash(password), False, activate_token)

        self.storage.insert_user(user)

        return user
