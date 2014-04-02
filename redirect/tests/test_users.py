import unittest
from mock import MagicMock
from .. users import Users
from .. storage import User
from .. redirectutil import hash
from .. restexceptions import RestException

class Request:
    def __init__(self, args):
        self.args = args

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

        with self.assertRaises(RestException):
            users.create_new_user(request)

    def test_create_user_existing_domain(self):
        users = Users(self.storage)
        self.storage.get_user_by_email = MagicMock(return_value=None)
        existing = User('boris', None, None, None, 'valid@mail.com', hash('pass123456'), True, None)
        self.storage.get_user_by_domain = MagicMock(return_value=existing)
        self.storage.insert_user = MagicMock()

        request = {'user_domain': 'boris', 'email': 'boris@mail.com', 'password': 'pass123456'}

        with self.assertRaises(RestException):
            users.create_new_user(request)
