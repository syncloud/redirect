import unittest

import smtp
from redirect.mail import Smtp, Mail
from redirect.models import User, ActionType
from redirect.services import Users
from redirect.servicesexceptions import ServiceException
from redirect.util import hash
from test.helpers import get_test_storage_creator


class TestUsers(unittest.TestCase):

    def setUp(self):
        self.activate_url_template = 'http://redirect.com?activate?token={0}'

        self.mail = Mail(Smtp('mail', 1025), 'support@redirect.com', self.activate_url_template, None, None)
        self.create_storage = get_test_storage_creator()
        with self.create_storage() as storage:
            storage.clear()

    def tearDown(self):
        smtp.clear()
        with self.create_storage() as storage:
            storage.clear()

    def add_user(self, user):
        with self.create_storage() as storage:
            storage.add(user)

    def get_user(self, email):
        with self.create_storage() as storage:
            return storage.get_user_by_email(email)

    def get_users_service(self, activate_by_email=True):
        return Users(self.create_storage, activate_by_email, self.mail)

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
