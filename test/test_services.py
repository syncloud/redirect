import unittest
from mock import MagicMock
from fakesmtp import FakeSmtp
from redirect.models import User, Action, ActionType
from redirect.util import hash
from redirect.mail import Smtp, Mail
from redirect.services import Users
from redirect.servicesexceptions import ServiceException
from test.helpers import get_test_storage_creator


class TestUsers(unittest.TestCase):

    def setUp(self):
        self.activate_url_template = 'http://redirect.com?activate?token={0}'

        self.mail = Mail(Smtp('localhost', 2500), 'support@redirect.com', self.activate_url_template, None)
        self.smtp = FakeSmtp('outbox', 'localhost', 2500)
        self.dns = MagicMock()
        self.create_storage = get_test_storage_creator()
        with self.create_storage() as storage:
            storage.clear()

    def tearDown(self):
        self.smtp.stop()
        with self.create_storage() as storage:
            storage.clear()

    def add_user(self, user):
        with self.create_storage() as storage:
            storage.add(user)

    def get_user(self, email):
        with self.create_storage() as storage:
            return storage.get_user_by_email(email)

    def get_users_service(self, activate_by_email=True):
        return Users(self.create_storage, activate_by_email, self.mail, self.dns, 'redirect.com')

    def test_user_create_success(self):
        users = self.get_users_service()

        request = {'email': u'valid@mail.com', 'password': u'pass123456'}
        user = users.create_new_user(request)

        self.assertIsNotNone(user)
        self.assertEqual('valid@mail.com', user.email)
        self.assertNotEqual('pass123456', user.password_hash, 'we should not store password plainly')
        self.assertIsNotNone(user.token(ActionType.ACTIVATE))
        self.assertFalse(user.active)

        activate_url = self.activate_url_template.format(user.token(ActionType.ACTIVATE))
        self.assertFalse(self.smtp.empty())
        email = self.smtp.emails()[0]
        self.assertTrue(user.email in email)
        self.assertTrue(activate_url in email)

    def test_user_create_no_activation(self):
        users = self.get_users_service(activate_by_email=False)

        request = {'email': u'valid@mail.com', 'password': u'pass123456'}
        user = users.create_new_user(request)

        self.assertIsNotNone(user)
        self.assertEqual('valid@mail.com', user.email)
        self.assertNotEqual('pass123456', user.password_hash, 'we should not store password plainly')
        # self.assertIsNone(user.activate_token())
        self.assertTrue(user.active)

        self.assertTrue(self.smtp.empty())

    def test_user_create_existing_email(self):
        users = self.get_users_service()
        existing = User(u'valid@mail.com', hash('pass123456'), True)
        self.add_user(existing)

        request = {'email': 'valid@mail.com', 'password': 'pass123456'}

        with self.assertRaises(ServiceException) as context:
            users.create_new_user(request)
        self.assertEquals(context.exception.status_code, 400)

    def test_user_create_missing_email(self):
        users = self.get_users_service()

        request = {'user_domain': 'boris', 'password': 'pass123456'}

        with self.assertRaises(ServiceException) as context:
            users.create_new_user(request)
        ex = context.exception
        self.assertEquals(ex.status_code, 400)
        self.assertGreater(len(ex.message), 0)
        self.assertEquals(len(ex.parameters_errors), 1)
        self.assertGreater(ex.parameters_errors['email'], 0)

    def test_user_activate_success(self):
        users = self.get_users_service()
        user = User(u'boris@mail.com', 'hash123', active=False)
        activate = user.enable_action(ActionType.ACTIVATE)
        self.add_user(user)

        request = {'token': activate.token}
        users.activate(request)

        user = self.get_user(user.email)
        self.assertTrue(user.active)

    def test_user_activate_missing_token(self):
        users = self.get_users_service()
        request = {}

        with self.assertRaises(ServiceException) as context:
            users.activate(request)

        exc = context.exception
        self.assertEquals(exc.status_code, 400)
        self.assertGreater(len(exc.message), 0)
        self.assertEquals(len(exc.parameters_errors), 1)
        self.assertGreater(len(exc.parameters_errors['token']), 0)

    def test_user_activate_wrong_token(self):
        users = self.get_users_service()

        request = {'token': 'wrong token 123'}

        with self.assertRaises(ServiceException) as context:
            users.activate(request)
        self.assertEquals(context.exception.status_code, 400)

    def test_user_activate_already_active(self):
        users = self.get_users_service()
        user = User(u'boris@mail.com', 'hash123', active=True)
        activate = user.enable_action(ActionType.ACTIVATE)
        self.add_user(user)

        request = {'token': activate.token}

        with self.assertRaises(ServiceException) as context:
            users.activate(request)
        self.assertEquals(context.exception.status_code, 400)

    def test_user_authenticate_success(self):
        users = self.get_users_service()
        user = User(u'boris@mail.com', hash('pass1234'), True)
        self.add_user(user)

        request = {'email': u'boris@mail.com', 'password': u'pass1234'}
        user = users.authenticate(request)

        self.assertIsNotNone(user)

    def test_user_authenticate_wrong_password(self):
        users = self.get_users_service()
        user = User(u'boris@mail.com', hash('otherpass1234'), True)
        self.add_user(user)

        request = {'email': 'boris@mail.com', 'password': 'pass1234'}
        with self.assertRaises(ServiceException) as context:
            users.authenticate(request)

        self.assertEquals(context.exception.status_code, 400)

    def test_user_authenticate_not_existing(self):
        users = self.get_users_service()

        request = {'email': 'boris@mail.com', 'password': 'pass1234'}
        with self.assertRaises(ServiceException) as context:
            users.authenticate(request)

        self.assertEquals(context.exception.status_code, 400)

    def test_user_authenticate_non_active(self):
        users = self.get_users_service()
        user = User(u'boris@mail.com', hash('pass1234'), active=False)
        user.enable_action(ActionType.ACTIVATE)
        self.add_user(user)

        request = {'email': 'boris@mail.com', 'password': 'pass1234'}
        with self.assertRaises(ServiceException) as context:
            users.authenticate(request)

        self.assertEquals(context.exception.status_code, 400)

    def test_user_authenticate_missing_password(self):
        users = self.get_users_service()
        user = User(u'boris@mail.com', hash('otherpass1234'), True)
        self.add_user(user)

        request = {'email': 'boris@mail.com'}
        with self.assertRaises(ServiceException) as context:
            users.authenticate(request)

        exc = context.exception
        self.assertEquals(exc.status_code, 400)
        self.assertGreater(len(exc.message), 0)
        self.assertEquals(len(exc.parameters_errors), 1)
        self.assertGreater(len(exc.parameters_errors['password']), 0)

    def test_user_set_subscribed_false(self):
        users = self.get_users_service()
        user = User(u'boris@mail.com', hash('pass'), True)
        self.add_user(user)

        request = {'subscribed': 'false'}
        users.user_set_subscribed(request, user.email)

        user = self.get_user(user.email)
        self.assertTrue(user.unsubscribed)

    def test_user_set_subscribed_true(self):
        users = self.get_users_service()
        user = User(u'boris@mail.com', hash('otherpass1234'), True)
        self.add_user(user)

        request = {'subscribed': 'false'}
        users.user_set_subscribed(request, user.email)

        request = {'subscribed': 'true'}
        users.user_set_subscribed(request, user.email)

        user = self.get_user(user.email)
        self.assertFalse(user.unsubscribed)
