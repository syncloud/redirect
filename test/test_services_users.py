import unittest
from mock import MagicMock
from redirect.models import User
from redirect.services import Users
from redirect.util import hash
from redirect.mail import Mail
from redirect.servicesexceptions import ServiceException
from fakesmtp import FakeSmtp

class TestUsers(unittest.TestCase):

    def setUp(self):
        self.storage = MagicMock()
        self.mail = Mail('localhost', 2500, 'support@redirect.com')
        self.smtp = FakeSmtp('outbox')
        self.smtp.clear()
        self.activate_url_template = 'http://redirect.com?activate?token={0}'
        self.dns = MagicMock()

    def get_users_service(self, activate_by_email=True):
        return Users(self.storage, activate_by_email, self.mail, self.activate_url_template, self.dns, 'redirect.com')

    def test_create_user_success(self):
        users = self.get_users_service()
        self.storage.get_user_by_email = MagicMock(return_value=None)
        self.storage.insert_user = MagicMock()

        request = {'user_domain': 'boris', 'email': 'valid@mail.com', 'password': 'pass123456'}
        user = users.create_new_user(request)

        self.assertTrue(self.storage.insert_user.called, 'user should be inserted in storage')

        self.assertIsNotNone(user)
        self.assertEqual('boris', user.user_domain)
        self.assertEqual('valid@mail.com', user.email)
        self.assertNotEqual('pass123456', user.password_hash, 'we should not store password plainly')
        self.assertIsNotNone(user.activate_token)
        self.assertFalse(user.active)

        activate_url = self.activate_url_template.format(user.activate_token)
        self.assertFalse(self.smtp.empty())
        email = self.smtp.emails()[0]
        self.assertTrue(user.email in email)
        self.assertTrue(activate_url in email)

    def test_create_user_no_activation(self):
        users = self.get_users_service(activate_by_email=False)
        self.storage.get_user_by_email = MagicMock(return_value=None)
        self.storage.insert_user = MagicMock()

        request = {'user_domain': 'boris', 'email': 'valid@mail.com', 'password': 'pass123456'}
        user = users.create_new_user(request)

        self.assertTrue(self.storage.insert_user.called, 'user should be inserted in storage')

        self.assertIsNotNone(user)
        self.assertEqual('boris', user.user_domain)
        self.assertEqual('valid@mail.com', user.email)
        self.assertNotEqual('pass123456', user.password_hash, 'we should not store password plainly')
        self.assertIsNone(user.activate_token)
        self.assertTrue(user.active)

        self.assertTrue(self.smtp.empty())

    def test_create_user_existing_email(self):
        users = self.get_users_service()
        existing = User('boris', None, None, None, 'valid@mail.com', hash('pass123456'), True, None)
        self.storage.get_user_by_email = MagicMock(return_value=existing)
        self.storage.insert_user = MagicMock()

        request = {'user_domain': 'vladimir', 'email': 'valid@mail.com', 'password': 'pass123456'}

        with self.assertRaises(ServiceException) as context:
            users.create_new_user(request)
        self.assertEquals(context.exception.status_code, 409)

    def test_create_user_existing_domain(self):
        users = self.get_users_service()
        self.storage.get_user_by_email = MagicMock(return_value=None)
        existing = User('boris', None, None, None, 'valid@mail.com', hash('pass123456'), True, None)
        self.storage.get_user_by_domain = MagicMock(return_value=existing)
        self.storage.insert_user = MagicMock()

        request = {'user_domain': 'boris', 'email': 'boris@mail.com', 'password': 'pass123456'}

        with self.assertRaises(ServiceException) as context:
            users.create_new_user(request)
        self.assertEquals(context.exception.status_code, 409)

    def test_create_user_missing_email(self):
        users = self.get_users_service()
        self.storage.get_user_by_email = MagicMock(return_value=None)
        self.storage.insert_user = MagicMock()

        request = {'user_domain': 'boris', 'password': 'pass123456'}

        with self.assertRaises(ServiceException) as context:
            users.create_new_user(request)
        self.assertEquals(context.exception.status_code, 400)
        self.assertGreater(len(context.exception.message), 0)

    def test_activate_success(self):
        users = self.get_users_service()
        user = User('boris', 'updatetoken123', None, None, 'boris@mail.com', 'hash123', False, 'activatetoken123')
        self.storage.get_user_by_activate_token = MagicMock(return_value=user)

        request = {'token': 'activatetoken123'}
        users.activate(request)

        self.assertTrue(user.active)
        self.assertTrue(self.storage.update_user.called)

    def test_activate_missing_token(self):
        users = self.get_users_service()
        request = {}

        with self.assertRaises(ServiceException) as context:
            users.activate(request)
        self.assertEquals(context.exception.status_code, 400)
        self.assertGreater(len(context.exception.message), 0)

    def test_activate_wrong_token(self):
        users = self.get_users_service()
        self.storage.get_user_by_activate_token = MagicMock(return_value=None)

        request = {'token': 'wrong token 123'}

        with self.assertRaises(ServiceException) as context:
            users.activate(request)
        self.assertEquals(context.exception.status_code, 400)

    def test_activate_already_active(self):
        users = self.get_users_service()
        user = User('boris', 'updatetoken123', None, None, 'boris@mail.com', 'hash123', True, 'activatetoken123')
        self.storage.get_user_by_activate_token = MagicMock(return_value=user)

        request = {'token': 'activatetoken123'}

        with self.assertRaises(ServiceException) as context:
            users.activate(request)
        self.assertEquals(context.exception.status_code, 409)

    def test_authenticate_success(self):
        users = self.get_users_service()
        user = User('boris', 'updatetoken123', None, None, 'boris@mail.com', hash('pass1234'), True, None)
        self.storage.get_user_by_email = MagicMock(return_value=user)

        request = {'email': 'boris@mail.com', 'password': 'pass1234'}
        user = users.authenticate(request)

        self.assertIsNotNone(user)

    def test_authenticate_wrong_password(self):
        users = self.get_users_service()
        user = User('boris', 'updatetoken123', None, None, 'boris@mail.com', hash('otherpass1234'), True, None)
        self.storage.get_user_by_email = MagicMock(return_value=user)

        request = {'email': 'boris@mail.com', 'password': 'pass1234'}
        with self.assertRaises(ServiceException) as context:
            users.authenticate(request)

        self.assertEquals(context.exception.status_code, 403)

    def test_authenticate_not_existing_user(self):
        users = self.get_users_service()
        self.storage.get_user_by_email = MagicMock(return_value=None)

        request = {'email': 'boris@mail.com', 'password': 'pass1234'}
        with self.assertRaises(ServiceException) as context:
            users.authenticate(request)

        self.assertEquals(context.exception.status_code, 403)

    def test_authenticate_non_active_user(self):
        users = self.get_users_service()
        user = User('boris', 'updatetoken123', None, None, 'boris@mail.com', hash('pass1234'), False, 'token123')
        self.storage.get_user_by_email = MagicMock(return_value=user)

        request = {'email': 'boris@mail.com', 'password': 'pass1234'}
        with self.assertRaises(ServiceException) as context:
            users.authenticate(request)

        self.assertEquals(context.exception.status_code, 403)

    def test_authenticate_missing_password(self):
        users = self.get_users_service()
        user = User('boris', 'updatetoken123', None, None, 'boris@mail.com', hash('otherpass1234'), True, None)
        self.storage.get_user_by_email = MagicMock(return_value=user)

        request = {'email': 'boris@mail.com'}
        with self.assertRaises(ServiceException) as context:
            users.authenticate(request)

        self.assertEquals(context.exception.status_code, 400)

    def test_update_success_new_ip(self):
        users = self.get_users_service()
        user = User('boris', 'updatetoken123', None, None, 'boris@mail.com', 'hash123', True, None)
        self.storage.get_user_by_update_token = MagicMock(return_value=user)

        request = {'token': 'updatetoken123', 'ip': '127.0.0.1', 'port': '10001'}
        user = users.update_ip_port(request)

        self.assertEquals('127.0.0.1', user.ip)
        self.assertEquals(10001, user.port)
        self.assertTrue(self.storage.update_user.called)
        self.assertTrue(self.dns.create_records.called)

    def test_update_success_update_ip(self):
        users = self.get_users_service()
        user = User('boris', 'updatetoken123', '127.0.0.1', 10001, 'boris@mail.com', 'hash123', True, None)
        self.storage.get_user_by_update_token = MagicMock(return_value=user)

        request = {'token': 'updatetoken123', 'ip': '127.0.0.2', 'port': '10002'}
        user = users.update_ip_port(request)

        self.assertEquals('127.0.0.2', user.ip)
        self.assertEquals(10002, user.port)
        self.assertTrue(self.storage.update_user.called)
        self.assertTrue(self.dns.update_records.called)

    def test_update_wrong_token(self):
        users = self.get_users_service()
        self.storage.get_user_by_update_token = MagicMock(return_value=None)

        request = {'token': 'updatetoken123', 'ip': '127.0.0.1', 'port': '10001'}

        with self.assertRaises(ServiceException) as context:
            users.update_ip_port(request)

        self.assertEquals(context.exception.status_code, 400)

    def test_update_missing_port(self):
        users = self.get_users_service()
        self.storage.get_user_by_update_token = MagicMock(return_value=None)

        request = {'token': 'updatetoken123', 'ip': '127.0.0.1'}

        with self.assertRaises(ServiceException) as context:
            users.update_ip_port(request)

        self.assertEquals(context.exception.status_code, 400)
        self.assertGreater(len(context.exception.message), 0)


    def test_update_non_active_user(self):
        users = self.get_users_service()
        user = User('boris', 'updatetoken123', None, None, 'boris@mail.com', 'hash123', False, 'token1234')
        self.storage.get_user_by_update_token = MagicMock(return_value=user)

        request = {'token': 'updatetoken123', 'ip': '127.0.0.1', 'port': '10001'}

        with self.assertRaises(ServiceException) as context:
            users.update_ip_port(request)

        self.assertEquals(context.exception.status_code, 400)
