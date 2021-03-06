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

    def create_new_user(self, request):
        validator = Validator(request)
        email = validator.email()
        password = validator.new_password()
        check_validator(validator)

        user = None
        action = None
        with self.create_storage() as storage:
            by_email = storage.get_user_by_email(email)
            if by_email and by_email.email == email:
                raise servicesexceptions.parameter_error('email', 'Email is already registered')

            user = User(email, util.hash(password), not self.activate_by_email)

            if self.activate_by_email:
                action = user.enable_action(ActionType.ACTIVATE)

            storage.add(user)

        if self.activate_by_email:
            self.mail.send_activate(self.main_domain, user.email, action.token)

        return user

    def activate(self, request):
        validator = Validator(request)
        token = validator.token()
        check_validator(validator)

        with self.create_storage() as storage:
            user = storage.get_user_by_activate_token(token)
            if not user:
                raise servicesexceptions.bad_request('Invalid activation token')

            if user.active:
                raise servicesexceptions.bad_request('User is active already')

            user.active = True

        return True

    def drop_device(self, request):
        self.authenticate(request)
        validator = Validator(request)
        user_domain = validator.new_user_domain()
        check_validator(validator)

        with self.create_storage() as storage:
            domain = storage.get_domain_by_name(user_domain)

            if not domain or not domain.user.active:
                raise servicesexceptions.bad_request('Unknown domain')

            domain.update_token = None
            domain.device_mac_address = None
            domain.device_name = None
            domain.device_title = None
            domain.ip = None
            domain.local_ip = None

            self.dns.delete_domain(self.main_domain, domain)

            return domain


    def domain_delete(self, request):
        user = self.authenticate(request)
        self.user_domain_delete(request, user)

    def user_domain_delete(self, request, user):
        validator = Validator(request)
        user_domain = validator.user_domain()
        check_validator(validator)

        with self.create_storage() as storage:
            domain = storage.get_domain_by_name(user_domain)

            if not domain or domain.user.email != user.email:
                raise servicesexceptions.bad_request('Unknown domain')

            self.dns.delete_domain(self.main_domain, domain)

            storage.delete_domain(domain)

    def user_set_subscribed(self, request, user_email):
        validator = Validator(request)
        subscribed = validator.boolean('subscribed', required=True)
        check_validator(validator)

        with self.create_storage() as storage:
            user = storage.get_user_by_email(user_email)
            if not user:
                raise servicesexceptions.bad_request('Unknown user')
            user.unsubscribed = not subscribed

    def delete_user(self, request):
        validator = Validator(request)
        email = validator.email()
        password = validator.password()
        check_validator(validator)

        with self.create_storage() as storage:
            user = storage.get_user_by_email(email)

            if not user or not user.active or not util.hash(password) == user.password_hash:
                raise servicesexceptions.bad_request('Authentication failed')

            for domain in user.domains:
                self.dns.delete_domain(self.main_domain, domain)

            storage.delete_user(user)

    def do_delete_user(self, email):
        with self.create_storage() as storage:
            user = storage.get_user_by_email(email)

            if not user:
                raise servicesexceptions.bad_request('Authentication failed')

            for domain in user.domains:
                self.dns.delete_domain(self.main_domain, domain)

            storage.delete_user(user)

    def do_user_domain_delete(self, user_domain):
        with self.create_storage() as storage:
            domain = storage.get_domain_by_name(user_domain)

            if not domain:
                raise servicesexceptions.bad_request('Unknown domain')

            self.dns.delete_domain(self.main_domain, domain)

            storage.delete_domain(domain)

    def user_reset_password(self, request):
        validator = Validator(request)
        email = validator.email()
        check_validator(validator)

        with self.create_storage() as storage:
            user = storage.get_user_by_email(email)

            if user and user.active:
                action = user.enable_action(ActionType.PASSWORD)

                self.mail.send_reset_password(user.email, action.token)

    def user_log(self, request):
        validator = Validator(request)
        token = validator.token()
        data = validator.string('data')
        include_support = validator.boolean('include_support', False, True)
        with self.create_storage() as storage:
            user = storage.get_user_by_update_token(token)
            if not user:
                raise servicesexceptions.bad_request('Invalid update token')
            self.mail.send_logs(user.email, data, include_support)

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

    def port_probe(self, request, request_ip):
        validator = Validator(request)
        token = validator.token()
        port = validator.port('port', True)
        protocol = validator.string('protocol', False)
        ip = validator.string('ip', False)
        check_validator(validator)
        domain = None
        with self.create_storage() as storage:
            domain = storage.get_domain_by_update_token(token)

        if not domain or not domain.user.active:
            raise servicesexceptions.bad_request('Unknown domain update token')

        try:
            if ip:
                request_ip = ip
            
            response = requests.get('{0}://{1}:{2}/ping'.format(protocol, request_ip, port),
                                    timeout=1, verify=False, allow_redirects=False)
            if response.status_code == 200:
                return {'message': response.text, 'device_ip': request_ip}, 200

        except Exception, e:
            pass

        return {'message': 'Port is not reachable', 'device_ip': request_ip}, 404
