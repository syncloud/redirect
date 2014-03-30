import uuid
from urlparse import urlparse


class AccountManager:
    def __init__(self, validation, database, dns, domain, token_by_mail, mail):
        self.mail = mail
        self.token_by_mail = token_by_mail
        self.domain = domain
        self.validation = validation
        self.database = database
        self.dns = dns

    def request_account(self, request):

        (errors, username, email, password, port, ip) = \
            self.validation.validate_create(request.args)

        if errors:
            return ", ".join(errors) + '\n', 400

        if port is None:
            port = request.remote_addr

        result = None
        status = None
        headers = {}

        exists = self.database.exists(username, email)

        if exists:
            result = 'User already exists'
            status = 409

        else:

            token = uuid.uuid4().hex

            try:
                self.database.insert(username, email, password, token, ip, port)

                result = "Created, check your mail for activation"
                status = 200
                if self.token_by_mail:
                    self.mail.send(username, email, token)
                else:
                    headers = {'Token': token}

            except Exception, e:
                result = 'Unable to create user: %s' % str(e)
                status = 500

        return result + '\n', status, headers

    def redirect_url(self, request, default_url):

        addr = urlparse(request.url).netloc

        url = default_url
        username = addr[:-(len(self.domain) + 1)]

        try:
            port = self.database.get_port_by_username(username)
            if port is not None:
                url = 'device.{0}.{1}:{2}/owncloud'.format(username, self.domain, port)
        except Exception:
            pass

        return 'http://' + url

    def activate(self, request):

        (errors, token) = self.validation.validate_token(request.args)
        if errors:
            return ", ".join(errors) + '\n', 400

        try:
            if self.database.activate(token):
                (username, ip, port) = self.database.get_user_info_by_token(token)
                self.dns.create_records(username, ip, port, self.domain)
                return "Activated\n", 200
            else:
                return "Not valid token\n", 400
        except Exception, e:
            return "Not activated: {0}\n".format(e), 500

    def token(self, request):

        (errors, username, password) = self.validation.validate_credentials(request.args)
        if errors:
            return ", ".join(errors) + '\n', 400

        try:

            if self.database.valid_user(username, password):
                token = self.database.get_token_by_password(username, password)
                return "Token found\n", 200, {'Token': token}
            else:
                return "User does not exist or password is incorrect\n", 400
        except Exception, e:
            return "Unable to get token: {0}\n".format(e), 500

    def update(self, request):

        (errors, token, new_ip, new_port) = self.validation.validate_update(request.args)
        if errors:
            return ", ".join(errors) + '\n', 400

        if new_port is None:
            new_port = request.remote_addr

        try:

            if self.database.existing_token(token):
                (username, ip, port) = self.database.get_user_info_by_token(token)

                if new_ip == ip and new_port == port:
                    return "No modified\n".format(ip, port), 304
                else:
                    self.dns.update_records(username, new_ip, new_port, self.domain)
                    self.database.update(token, new_ip, new_port)

                    return "Updated to {0}:{1}\n".format(new_ip, new_port), 200
            else:
                return "Not valid token\n", 400
        except Exception, e:
            return "Unable to update: {0}\n".format(e), 500

    def delete(self, request):

        (errors, username, password) = self.validation.validate_credentials(request.args)
        if errors:
            return ", ".join(errors) + '\n', 400

        try:

            if self.database.valid_user(username, password):
                (username, ip, port) = self.database.get_user_info_by_password(username, password)
                self.dns.delete_records(username, ip, port, self.domain)
                self.database.delete_user(username, password)
                return "User and dns are removed\n", 200
            else:
                return "User does not exist or password is incorrect\n", 400
        except Exception, e:
            return "Unable to update: {0}\n".format(e), 500
