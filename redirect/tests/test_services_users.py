import unittest
from mock import MagicMock
from redirect.models import User
from .. services import Users
from .. util import hash
from .. servicesexceptions import ServiceException

class TestCreateLogin(unittest.TestCase):

    def setUp(self):
        self.storage = MagicMock()

    def test_create_user_success(self):
        users = Users(self.storage)
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

    def test_create_user_existing_email(self):
        users = Users(self.storage)
        existing = User('boris', None, None, None, 'valid@mail.com', hash('pass123456'), True, None)
        self.storage.get_user_by_email = MagicMock(return_value=existing)
        self.storage.insert_user = MagicMock()

        request = {'user_domain': 'vladimir', 'email': 'valid@mail.com', 'password': 'pass123456'}

        with self.assertRaises(ServiceException) as context:
            users.create_new_user(request)
        self.assertEquals(context.exception.status_code, 409)

    def test_create_user_existing_domain(self):
        users = Users(self.storage)
        self.storage.get_user_by_email = MagicMock(return_value=None)
        existing = User('boris', None, None, None, 'valid@mail.com', hash('pass123456'), True, None)
        self.storage.get_user_by_domain = MagicMock(return_value=existing)
        self.storage.insert_user = MagicMock()

        request = {'user_domain': 'boris', 'email': 'boris@mail.com', 'password': 'pass123456'}

        with self.assertRaises(ServiceException) as context:
            users.create_new_user(request)
        self.assertEquals(context.exception.status_code, 409)

    def test_create_user_missing_email(self):
        users = Users(self.storage)
        self.storage.get_user_by_email = MagicMock(return_value=None)
        self.storage.insert_user = MagicMock()

        request = {'user_domain': 'boris', 'password': 'pass123456'}

        with self.assertRaises(ServiceException) as context:
            users.create_new_user(request)
        self.assertEquals(context.exception.status_code, 400)
        self.assertGreater(len(context.exception.message), 0)

    def test_activate_success(self):
        users = Users(self.storage)
        user = User('boris', 'updatetoken123', None, None, 'boris@mail.com', 'hash123', False, 'activatetoken123')
        self.storage.get_user_by_activate_token = MagicMock(return_value=user)

        request = {'token': 'activatetoken123'}
        success = users.activate(request)

        self.assertTrue(success)
        self.assertTrue(user.active)
        self.assertTrue(self.storage.update_user.called)

    def test_activate_missing_token(self):
        users = Users(self.storage)
        user = User('boris', 'updatetoken123', None, None, 'boris@mail.com', 'hash123', False, 'activatetoken123')
        self.storage.get_user_by_activate_token = MagicMock(return_value=user)

        request = {}

        with self.assertRaises(ServiceException) as context:
            users.activate(request)
        self.assertEquals(context.exception.status_code, 400)
        self.assertGreater(len(context.exception.message), 0)

    def test_activate_wrong_token(self):
        users = Users(self.storage)
        self.storage.get_user_by_activate_token = MagicMock(return_value=None)

        request = {'token': 'wrong token 123'}

        with self.assertRaises(ServiceException) as context:
            users.activate(request)
        self.assertEquals(context.exception.status_code, 400)