import unittest
from mock import MagicMock
from fakesmtp import FakeSmtp
from helpers import get_storage_creator

from redirect.models import User
from redirect.util import hash
from redirect.mail import Mail
from redirect.services import Users
from redirect.servicesexceptions import ServiceException

from redirect.storage import Storage

class TestUsers(unittest.TestCase):

    def setUp(self):
        self.mail = Mail('localhost', 2500, 'support@redirect.com')
        self.smtp = FakeSmtp('outbox')
        self.smtp.clear()
        self.activate_url_template = 'http://redirect.com?activate?token={0}'
        self.dns = MagicMock()
        self.create_storage = get_storage_creator()

    def tearDown(self):
        with self.create_storage() as session:
            storage = Storage(session)
            storage.clear()

    def add_user(self, user):
        with self.create_storage() as session:
            storage = Storage(session)
            storage.add(user)

    def get_user(self, email):
        with self.create_storage() as session:
            storage = Storage(session)
            return storage.get_user_by_email(email)

    def get_users_service(self, activate_by_email=True):
        return Users(self.create_storage, activate_by_email, self.mail, self.activate_url_template, self.dns, 'redirect.com')

    def test_user_create_success(self):
        users = self.get_users_service()

        request = {'user_domain': u'boris', 'email': u'valid@mail.com', 'password': u'pass123456'}
        user = users.create_new_user(request)

        self.assertIsNotNone(user)
        self.assertEqual('valid@mail.com', user.email)
        self.assertNotEqual('pass123456', user.password_hash, 'we should not store password plainly')
        self.assertIsNotNone(user.activate_token)
        self.assertFalse(user.active)

        self.assertEquals(1, len(user.domains))
        self.assertEqual('boris', user.domains[0].user_domain)

        activate_url = self.activate_url_template.format(user.activate_token)
        self.assertFalse(self.smtp.empty())
        email = self.smtp.emails()[0]
        self.assertTrue(user.email in email)
        self.assertTrue(activate_url in email)

    def test_user_create_no_activation(self):
        users = self.get_users_service(activate_by_email=False)

        request = {'user_domain': u'boris', 'email': u'valid@mail.com', 'password': u'pass123456'}
        user = users.create_new_user(request)

        self.assertIsNotNone(user)
        self.assertEqual('valid@mail.com', user.email)
        self.assertNotEqual('pass123456', user.password_hash, 'we should not store password plainly')
        self.assertIsNone(user.activate_token)
        self.assertTrue(user.active)

        self.assertEquals(1, len(user.domains))
        self.assertEqual('boris', user.domains[0].user_domain)

        self.assertTrue(self.smtp.empty())

    def test_user_create_existing_email(self):
        users = self.get_users_service()
        existing = User(u'valid@mail.com', hash('pass123456'), True, None)

        self.add_user(existing)

        request = {'user_domain': 'vladimir', 'email': 'valid@mail.com', 'password': 'pass123456'}

        with self.assertRaises(ServiceException) as context:
            users.create_new_user(request)
        self.assertEquals(context.exception.status_code, 409)

    def test_user_create_existing_domain(self):
        users = self.get_users_service()

        request = {'user_domain': 'boris', 'email': 'boris@mail.com', 'password': 'pass123456'}

        users.create_new_user(request)

        with self.assertRaises(ServiceException) as context:
            users.create_new_user(request)
        self.assertEquals(context.exception.status_code, 409)

    def test_user_create_missing_email(self):
        users = self.get_users_service()

        request = {'user_domain': 'boris', 'password': 'pass123456'}

        with self.assertRaises(ServiceException) as context:
            users.create_new_user(request)
        self.assertEquals(context.exception.status_code, 400)
        self.assertGreater(len(context.exception.message), 0)

    def test_user_activate_success(self):
        users = self.get_users_service()
        user = User(u'boris@mail.com', 'hash123', False, u'activatetoken123')

        self.add_user(user)

        request = {'token': u'activatetoken123'}
        users.activate(request)

        user = self.get_user(user.email)
        self.assertTrue(user.active)

    def test_user_activate_missing_token(self):
        users = self.get_users_service()
        request = {}

        with self.assertRaises(ServiceException) as context:
            users.activate(request)
        self.assertEquals(context.exception.status_code, 400)
        self.assertGreater(len(context.exception.message), 0)

    def test_user_activate_wrong_token(self):
        users = self.get_users_service()

        request = {'token': 'wrong token 123'}

        with self.assertRaises(ServiceException) as context:
            users.activate(request)
        self.assertEquals(context.exception.status_code, 400)

    def test_user_activate_already_active(self):
        users = self.get_users_service()
        user = User(u'boris@mail.com', 'hash123', True, u'activatetoken123')
        self.add_user(user)

        request = {'token': u'activatetoken123'}

        with self.assertRaises(ServiceException) as context:
            users.activate(request)
        self.assertEquals(context.exception.status_code, 409)

    def test_user_authenticate_success(self):
        users = self.get_users_service()
        user = User(u'boris@mail.com', hash('pass1234'), True, None)
        self.add_user(user)

        request = {'email': u'boris@mail.com', 'password': u'pass1234'}
        user = users.authenticate(request)

        self.assertIsNotNone(user)

    def test_user_authenticate_wrong_password(self):
        users = self.get_users_service()
        user = User(u'boris@mail.com', hash('otherpass1234'), True, None)
        self.add_user(user)

        request = {'email': 'boris@mail.com', 'password': 'pass1234'}
        with self.assertRaises(ServiceException) as context:
            users.authenticate(request)

        self.assertEquals(context.exception.status_code, 403)

    def test_user_authenticate_not_existing(self):
        users = self.get_users_service()

        request = {'email': 'boris@mail.com', 'password': 'pass1234'}
        with self.assertRaises(ServiceException) as context:
            users.authenticate(request)

        self.assertEquals(context.exception.status_code, 403)

    def test_user_authenticate_non_active(self):
        users = self.get_users_service()
        user = User(u'boris@mail.com', hash('pass1234'), False, u'token123')
        self.add_user(user)

        request = {'email': 'boris@mail.com', 'password': 'pass1234'}
        with self.assertRaises(ServiceException) as context:
            users.authenticate(request)

        self.assertEquals(context.exception.status_code, 403)

    def test_user_authenticate_missing_password(self):
        users = self.get_users_service()
        user = User(u'boris@mail.com', hash('otherpass1234'), True, None)
        self.add_user(user)

        request = {'email': 'boris@mail.com'}
        with self.assertRaises(ServiceException) as context:
            users.authenticate(request)

        self.assertEquals(context.exception.status_code, 400)

    def test_domain_update_success_new_ip(self):
        users = self.get_users_service()

        request = {'user_domain': u'boris', 'email': u'valid@mail.com', 'password': u'pass123456'}
        user = users.create_new_user(request)

        request = {'token': user.activate_token}
        users.activate(request)

        domain = user.domains[0]

        request = {'token': domain.update_token, 'ip': u'127.0.0.1', 'port': u'10001'}
        updated = users.update_ip_port(request)

        self.assertEquals(1, len(updated.services))

        service = updated.services[0]

        self.assertEquals('127.0.0.1', updated.ip)
        self.assertEquals(10001, service.port)
        self.assertTrue(self.dns.create_records.called)

    def test_domain_update_success(self):
        users = self.get_users_service()

        request = {'user_domain': u'boris', 'email': u'valid@mail.com', 'password': u'pass123456'}
        user = users.create_new_user(request)

        request = {'token': user.activate_token}
        users.activate(request)

        domain = user.domains[0]

        request = {'token': domain.update_token, 'ip': u'127.0.0.1', 'port': u'10001'}
        updated = users.update_ip_port(request)

        request = {'token': domain.update_token, 'ip': u'127.0.0.2', 'port': u'10001'}
        updated = users.update_ip_port(request)

        service = updated.services[0]

        self.assertEquals('127.0.0.2', updated.ip)
        self.assertEquals(10001, service.port)
        self.assertTrue(self.dns.create_records.called)

    def test_domain_update_wrong_token(self):
        users = self.get_users_service()

        request = {'token': 'updatetoken123', 'ip': '127.0.0.1', 'port': '10001'}

        with self.assertRaises(ServiceException) as context:
            users.update_ip_port(request)

        self.assertEquals(context.exception.status_code, 400)

    def test_domain_update_missing_port(self):
        users = self.get_users_service()

        request = {'token': 'updatetoken123', 'ip': '127.0.0.1'}

        with self.assertRaises(ServiceException) as context:
            users.update_ip_port(request)

        self.assertEquals(context.exception.status_code, 400)
        self.assertGreater(len(context.exception.message), 0)


    def test_domain_update_non_active_user(self):
        users = self.get_users_service()

        request = {'user_domain': u'boris', 'email': u'valid@mail.com', 'password': u'pass123456'}
        user = users.create_new_user(request)

        domain = user.domains[0]

        request = {'token': domain.update_token, 'ip': u'127.0.0.1', 'port': u'10001'}

        with self.assertRaises(ServiceException) as context:
            users.update_ip_port(request)

        self.assertEquals(context.exception.status_code, 400)