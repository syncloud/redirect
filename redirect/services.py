import logging
from datetime import datetime

from models import User, Domain, new_service_from_dict, ActionType
from validation import Validator
import servicesexceptions
import util


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
        errors = validator.errors

        if errors:
            message = ", ".join(errors)
            raise servicesexceptions.bad_request(message)

        user = self.get_user(email)
        if not user or not user.active or not util.hash(password) == user.password_hash:
            raise servicesexceptions.forbidden('Authentication failed')

        return user

    def redirect_url(self, request_url):
        user_domain = util.get_second_level_domain(request_url, self.main_domain)

        if not user_domain:
            raise servicesexceptions.bad_request('Second level domain should be specified')

        user = self.storage.get_user_by_domain(user_domain)

        if not user:
            raise servicesexceptions.not_found('The second level domain is not registered')

        return 'http://device.{0}.{1}:{2}/owncloud'.format(user_domain, self.main_domain, user.port)


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
        errors = validator.errors

        if errors:
            message = ", ".join(errors)
            raise servicesexceptions.bad_request(message)

        user = self.get_user(email)
        if not user or not user.active or not util.hash(password) == user.password_hash:
            raise servicesexceptions.forbidden('Authentication failed')

        return user

    def create_new_user(self, request):
        validator = Validator(request)
        email = validator.email()
        password = validator.password()
        user_domain = validator.new_user_domain(error_if_missing=False)
        errors = validator.errors

        if errors:
            message = ", ".join(errors)
            raise servicesexceptions.bad_request(message)

        user = None
        action = None
        with self.create_storage() as storage:
            by_email = storage.get_user_by_email(email)
            if by_email and by_email.email == email:
                raise servicesexceptions.conflict('Email is already registered')

            if user_domain:
                by_domain = storage.get_domain_by_name(user_domain)
                if by_domain and by_domain.user_domain == user_domain:
                    raise servicesexceptions.conflict('User domain name is already in use')

            update_token = util.create_token()

            user = User(email, util.hash(password), not self.activate_by_email)

            if user_domain:
                domain = Domain(user_domain, None, update_token)
                domain.user = user
                user.domains.append(domain)
                storage.add(domain)

            if self.activate_by_email:
                action = user.enable_action(ActionType.ACTIVATE)

            storage.add(user)

        if self.activate_by_email:
            self.mail.send_activate(self.main_domain, user.email, action.token)

        return user

    def activate(self, request):
        validator = Validator(request)
        token = validator.token()
        errors = validator.errors

        if errors:
            message = ", ".join(errors)
            raise servicesexceptions.bad_request(message)

        with self.create_storage() as storage:
            user = storage.get_user_by_activate_token(token)
            if not user:
                raise servicesexceptions.bad_request('Invalid activation token')

            if user.active:
                raise servicesexceptions.conflict('User is active already')

            user.active = True

        return True

    def domain_acquire(self, request):
        user = self.authenticate(request)

        validator = Validator(request)
        user_domain = validator.new_user_domain()

        with self.create_storage() as storage:
            domain = storage.get_domain_by_name(user_domain)
            if domain and domain.user_id != user.id:
                raise servicesexceptions.conflict('User domain name is already in use')

            update_token = util.create_token()
            if not domain:
                domain = Domain(user_domain, None, update_token)
                domain.user = user
                storage.add(domain)
            else:
                domain.update_token = update_token

            return domain

    def service_compare(self, a, b):
        return a.port == b.port and a.type == b.type

    def get_missing(self, lookfor, lookat):
        result = []
        for s in lookfor:
            existing = None
            for x in lookat:
                if self.service_compare(x, s):
                    existing = x
            if not existing:
                result.append(s)
        return result

    def validate_service(self, data):
        validator = Validator(data)
        validator.port()
        if validator.errors:
            message = ", ".join(validator.errors)
            raise servicesexceptions.bad_request(message)

    def domain_update(self, request, request_ip=None):
        validator = Validator(request)
        token = validator.token()
        ip = validator.ip(request_ip)
        errors = validator.errors

        if errors:
            message = ", ".join(errors)
            raise servicesexceptions.bad_request(message)

        with self.create_storage() as storage:
            domain = storage.get_domain_by_update_token(token)

            if not domain or not domain.user.active:
                raise servicesexceptions.bad_request('Unknown domain update token')

            map(self.validate_service, request['services'])

            request_services = [new_service_from_dict(s) for s in request['services']]
            added_services = self.get_missing(request_services, domain.services)
            removed_services = self.get_missing(domain.services, request_services)

            storage.delete(removed_services)

            for s in added_services:
                s.domain = domain
                domain.services.append(s)

            storage.add(added_services)

            is_new_dmain = domain.ip is None
            update_ip = domain.ip != ip
            domain.ip = ip

            if is_new_dmain:
                self.dns.new_domain(self.main_domain, domain)
            else:
                self.dns.update_domain(self.main_domain, domain, update_ip=update_ip, added=added_services, removed=removed_services)

            domain.last_update = datetime.now()
            return domain

    def get_domain(self, request):
        validator = Validator(request)
        token = validator.token()
        with self.create_storage() as storage:
            domain = storage.get_domain_by_update_token(token)
            if not domain or not domain.user.active:
                raise servicesexceptions.bad_request('Unknown domain update token')
            return domain

    def delete_user(self, request):
        validator = Validator(request)
        email = validator.email()
        password = validator.password()
        errors = validator.errors

        if errors:
            message = ", ".join(errors)
            raise servicesexceptions.bad_request(message)

        with self.create_storage() as storage:
            user = storage.get_user_by_email(email)

            if not user or not user.active or not util.hash(password) == user.password_hash:
                raise servicesexceptions.forbidden('Authentication failed')

            for domain in user.domains:
                self.dns.delete_domain(self.main_domain, domain)

            storage.delete_user(user)

    def user_reset_password(self, request):
        validator = Validator(request)
        email = validator.email()
        errors = validator.errors

        if errors:
            message = ", ".join(errors)
            raise servicesexceptions.bad_request(message)

        with self.create_storage() as storage:
            user = storage.get_user_by_email(email)

            if user and user.active:
                action = user.enable_action(ActionType.PASSWORD)

                self.mail.send_reset_password(user.email, action.token)

    def user_set_password(self, request):
        validator = Validator(request)
        token = validator.token()
        password = validator.new_password()
        errors = validator.errors

        if errors:
            message = ", ".join(errors)
            raise servicesexceptions.bad_request(message)

        with self.create_storage() as storage:
            user = storage.get_user_by_token(ActionType.PASSWORD, token)

            if not user:
                raise servicesexceptions.forbidden('Invalid password token')

            user.password_hash = util.hash(password)

            self.mail.send_set_password(user.email)