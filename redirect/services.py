from models import User, Domain, Service, new_service, new_service_fromdict
from validation import Validator
import servicesexceptions
import util

class Users:
    def __init__(self, create_storage, activate_by_email, mail, activate_url_template, dns, domain):
        self.storage = None
        self.activate_by_email = activate_by_email
        self.mail = mail
        self.activate_url_template = activate_url_template
        self.dns = dns
        self.main_domain = domain
        self.create_storage = create_storage

    def get_user(self, email):
        with self.create_storage() as storage:
            return storage.get_user_by_email(email)

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
        domain = None
        with self.create_storage() as storage:
            by_email = storage.get_user_by_email(email)
            if by_email and by_email.email == email:
                raise servicesexceptions.conflict('Email is already registered')

            if user_domain:
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

            if user_domain:
                domain = Domain(user_domain, None, update_token)
                domain.user = user
                user.domains.append(domain)
                storage.add(domain)

            storage.add(user)

        if self.activate_by_email:
            activate_url = self.activate_url_template.format(user.activate_token)
            self.mail.send_activate(user_domain, self.main_domain, user.email, activate_url)

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


    def update_ip_port(self, request, request_ip=None):
        validator = Validator(request)
        token = validator.token()
        ip = validator.ip(request_ip)
        port = validator.port()
        errors = validator.errors

        if errors:
            message = ", ".join(errors)
            raise servicesexceptions.bad_request(message)

        with self.create_storage() as storage:
            domain = storage.get_domain_by_update_token(token)

            if not domain or not domain.user.active:
                raise servicesexceptions.bad_request('Unknown update token')

            if len(domain.services) > 0:
                service = domain.services[0]
                service.port = port
            else:
                service = new_service('owncloud', '_http._tcp', port)
                service.domain = domain
                domain.services.append(service)

            if domain.ip:
                self.dns.update_records(domain.user_domain, ip, port, self.main_domain)
            else:
                self.dns.create_records(domain.user_domain, ip, port, self.main_domain)

            domain.ip = ip
            return domain

    def get_missing(self, lookfor, lookat, compare):
        result = []
        for s in lookfor:
            existing = None
            for x in lookat:
                if compare(x, s):
                    existing = x
            if not existing:
                result.append(s)
        return result

    # def get_new_services(self, services, existing_services):
    #     return [s for s in services if not next(x for x in iter(existing_services) if x.port == s.port)]

    def update2(self, request, request_ip=None):
        validator = Validator(request)
        token = validator.token()
        ip = validator.ip(request_ip)
        errors = validator.errors

        if errors:
            message = ", ".join(errors)
            raise servicesexceptions.bad_request(message)

        service_compare = lambda a, b: (a.port == b.port)

        with self.create_storage() as storage:
            domain = storage.get_domain_by_update_token(token)

            if not domain or not domain.user.active:
                raise servicesexceptions.bad_request('Unknown update token')

            request_services = [new_service_fromdict(s) for s in request['services']]
            new_services = self.get_missing(request_services, domain.services, service_compare)
            removed_services = self.get_missing(domain.services, request_services, service_compare)

            for s in new_services:
                s.domain = domain
                domain.services.append(s)
                storage.add(s)

            is_new_dmain = domain.ip is None
            domain.ip = ip

            if is_new_dmain:
                self.dns.new_domain(self.main_domain, domain)
            else:
                self.dns.update_domain(self.main_domain, domain, update_ip=True, created=new_services, changed=[], removed=removed_services)

            return domain

    def redirect_url(self, request_url):

        user_domain = util.get_second_level_domain(request_url, self.main_domain)

        if not user_domain:
            raise servicesexceptions.bad_request('Second level domain should be specified')

        user = self.storage.get_user_by_domain(user_domain)

        if not user:
            raise servicesexceptions.not_found('The second level domain is not registered')

        return 'http://device.{0}.{1}:{2}/owncloud'.format(user_domain, self.main_domain, user.port)

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

        self.dns.delete_records(user.user_domain, user.ip, user.port, self.main_domain)

        return True