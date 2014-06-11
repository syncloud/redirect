from models import User
from validation import Validator
import servicesexceptions
import util

class Users:
    def __init__(self, user_storage, activate_by_email, mail, activate_url_template, dns, domain):
        self.storage = user_storage
        self.activate_by_email = activate_by_email
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

        update_token = util.create_token()
        activate_token = None
        active = True
        if self.activate_by_email:
            active = False
            activate_token = util.create_token()
        user = User(user_domain, update_token, None, None, email, util.hash(password), active, activate_token)

        self.storage.insert_user(user)
        self.storage.save()

        if self.activate_by_email:
            activate_url = self.activate_url_template.format(user.activate_token)
            full_domain = '{0}.{1}'.format(user.user_domain, self.domain)
            self.mail.send_activate(full_domain, user.email, activate_url)

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
        self.storage.save()

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
        if not user or not user.active or not util.hash(password) == user.password_hash:
            raise servicesexceptions.forbidden('Authentication failed')

        return user

    def update_ip_port(self, request, request_ip=None):
        validator = Validator(request)
        token = validator.token()
        ip = validator.ip(request_ip)
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
        self.storage.save()

        return user

    def redirect_url(self, request_url):

        user_domain = util.get_second_level_domain(request_url, self.domain)

        if not user_domain:
            raise servicesexceptions.bad_request('Second level domain should be specified')

        user = self.storage.get_user_by_domain(user_domain)

        if not user:
            raise servicesexceptions.not_found('The second level domain is not registered')

        return 'http://device.{0}.{1}:{2}/owncloud'.format(user_domain, self.domain, user.port)

    def delete_user(self, request):
        validator = Validator(request)
        email = validator.email()
        password = validator.password()
        errors = validator.errors

        if errors:
            message = ", ".join(errors)
            raise servicesexceptions.bad_request(message)

        user = self.get_user(email)
        if not user or not user.active or not util.hash(password) == user.password_hash:
            raise servicesexceptions.forbidden('Authentication failed')

        deleted = self.storage.delete_user(email)
        self.storage.save()
        if not deleted:
            raise servicesexceptions.conflict('Unable to delete user')

        self.dns.delete_records(user.user_domain, user.ip, user.port, self.domain)

        return True