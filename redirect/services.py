from models import User, Domain, Service
from validation import Validator
import servicesexceptions
import util
from storage import Storage

class Users:
    def __init__(self, create_storage, activate_by_email, mail, activate_url_template, dns, domain):
        self.storage = None
        self.activate_by_email = activate_by_email
        self.mail = mail
        self.activate_url_template = activate_url_template
        self.dns = dns
        self.domain = domain
        self.create_storage = create_storage

    def get_user(self, email):
        with self.create_storage() as session:
            return Storage(session).get_user_by_email(email)

    def create_new_user(self, request):
        validator = Validator(request)
        email = validator.email()
        password = validator.password()
        user_domain = validator.new_user_domain()
        errors = validator.errors

        if errors:
            message = ", ".join(errors)
            raise servicesexceptions.bad_request(message)

        user = None
        domain = None
        with self.create_storage() as session:
            storage = Storage(session)

            by_email = storage.get_user_by_email(email)
            if by_email and by_email.email == email:
                raise servicesexceptions.conflict('Email is already registered')

            by_domain = storage.get_domain_by_name(user_domain)
            if by_domain and by_domain.user_domain == user_domain:
                raise servicesexceptions.conflict('User domain name is already in use')

            update_token = util.create_token()
            activate_token = None
            active = True
            if self.activate_by_email:
                active = False
                activate_token = util.create_token()

            user = User(email, util.hash(password), active, activate_token)
            domain = Domain(user_domain, None, update_token)

            domain.user = user
            user.domains.append(domain)

            service = Service('owncloud', '_http._tcp', None)

            service.domain = domain
            domain.services.append(service)

            storage.add(user)
            storage.add(domain)
            storage.add(service)

        if self.activate_by_email:
            activate_url = self.activate_url_template.format(user.activate_token)
            full_domain = '{0}.{1}'.format(domain.user_domain, self.domain)
            self.mail.send_activate(full_domain, user.email, activate_url)

        return user

    def activate(self, request):
        validator = Validator(request)
        token = validator.token()
        errors = validator.errors

        if errors:
            message = ", ".join(errors)
            raise servicesexceptions.bad_request(message)

        with self.create_storage() as session:
            storage = Storage(session)

            user = storage.get_user_by_activate_token(token)
            if not user:
                raise servicesexceptions.bad_request('Invalid activation token')

            if user.active:
                raise servicesexceptions.conflict('User is active already')

            user.update_active(True)

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

        domain = None

        with self.create_storage() as session:
            storage = Storage(session)

            domain = storage.get_domain_by_update_token(token)

            if not domain or not domain.user.active:
                raise servicesexceptions.bad_request('Unknown update token')

            service = domain.services[0]

            if domain.ip:
                self.dns.update_records(domain.user_domain, ip, port, self.domain)
            else:
                self.dns.create_records(domain.user_domain, ip, port, self.domain)

            domain.ip = ip
            service.port = port

        return domain

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