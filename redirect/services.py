import uuid
from models import User
from validation import Validator
from util import hash
import servicesexceptions
from urlparse import urlparse
import util

class Users:
    def __init__(self, user_storage, mail, activate_url_template, dns, domain):
        self.storage = user_storage
        self.mail = mail
        self.activate_url_template = activate_url_template
        self.dns = dns
        self.domain = domain

    def get_user(self, email):
        return self.storage.get_user_by_email(email)

    def create_new_user(self, request):
        validator = Validator(request)
        email = validator.email()
        password = validator.password()
        user_domain = validator.new_user_domain()
        errors = validator.errors

        if errors:
            message = ", ".join(errors)
            raise servicesexceptions.bad_request(message)

        by_email = self.get_user(email)
        if by_email and by_email.email == email:
            raise servicesexceptions.conflict('Email is already registered')

        by_domain = self.storage.get_user_by_domain(user_domain)
        if by_domain and by_domain.user_domain == user_domain:
            raise servicesexceptions.conflict('User domain name is already in use')

        update_token = uuid.uuid4().hex
        activate_token = uuid.uuid4().hex
        user = User(user_domain, update_token, None, None, email, hash(password), False, activate_token)

        self.storage.insert_user(user)

        activate_url = self.activate_url_template.format(user.activate_token)
        self.mail.send(user.user_domain, user.email, activate_url)

        return user

    def activate(self, request):
        validator = Validator(request)
        token = validator.token()
        errors = validator.errors

        if errors:
            message = ", ".join(errors)
            raise servicesexceptions.bad_request(message)

        user = self.storage.get_user_by_activate_token(token)
        if not user:
            raise servicesexceptions.bad_request('Invalid activation token')

        if user.active:
            raise servicesexceptions.conflict('User is active already')

        user.update_active(True)
        self.storage.update_user(user)

        return True

    def authenticate(self, request):
        validator = Validator(request)
        email = validator.email()
        password = validator.password()
        errors = validator.errors

        if errors:
            message = ", ".join(errors)
            raise servicesexceptions.bad_request(message)

        user = self.get_user(email)
        if not user or not user.active or not hash(password) == user.password_hash:
            raise servicesexceptions.forbidden('Authentication failed')

        return user

    def get_update_token(self, request):
        user = self.authenticate(request)

        validator = Validator(request)
        user_domain = validator.user_domain()
        errors = validator.errors

        if errors:
            message = ", ".join(errors)
            raise servicesexceptions.bad_request(message)

        if not user_domain == user.user_domain:
            raise servicesexceptions.forbidden('Authentication failed')

        return user.update_token

    def update_ip_port(self, request):
        validator = Validator(request)
        token = validator.token()
        ip = validator.ip()
        port = validator.port()
        errors = validator.errors

        if errors:
            message = ", ".join(errors)
            raise servicesexceptions.bad_request(message)

        user = self.storage.get_user_by_update_token(token)

        if not user or not user.active:
            raise servicesexceptions.bad_request('Unknown update token')

        if user.ip:
            self.dns.update_records(user.user_domain, ip, port, self.domain)
        else:
            self.dns.create_records(user.user_domain, ip, port, self.domain)

        user.update_ip_port(ip, port)
        self.storage.update_user(user)

        return user

    def redirect_url(self, request_url):

        user_domain = util.get_second_level_domain(request_url, self.domain)

        if not user_domain:
            raise servicesexceptions.bad_request('Second level domain should be specified')

        user = self.storage.get_user_by_domain(user_domain)

        if not user:
            raise servicesexceptions.not_found('The second level domain is not registered')

        return 'http://device.{0}.{1}:{2}/owncloud'.format(user_domain, self.domain, user.port)