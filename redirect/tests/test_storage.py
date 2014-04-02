import uuid
import unittest
import ConfigParser
import os
from .. storage import User, UserStorage

class TestStorageUser(unittest.TestCase):

    def assertUser(self, expected, actual):

        if expected is None:
            self.assertIsNone(actual)
        if expected is not None:
            self.assertIsNotNone(actual)
        if expected is not None and actual is not None:
            self.assertEquals(expected.email, actual.email, 'Users should have the same email')
            self.assertEquals(expected.password_hash, actual.password_hash, 'Users should have the same password_hash')
            self.assertEquals(expected.active, actual.active, 'Users should have the same active')
            self.assertEquals(expected.user_domain, actual.user_domain, 'Users should have the same user_domain')
            self.assertEquals(expected.update_token, actual.update_token, 'Users should have the same update_token')
            self.assertEquals(expected.activate_token, actual.activate_token, 'Users should have the same activate_token')

    def generate_user(self):

        domain = uuid.uuid4().hex
        email = domain + '@mail.com'
        update_token = uuid.uuid4().hex
        activate_token = uuid.uuid4().hex
        user = User(domain, update_token, '127.0.0.1', 10001, email, 'hash1234', False, activate_token)
        return user


    def setUp(self):

        config = ConfigParser.ConfigParser()
        config_path = os.path.join(os.path.dirname(__file__), 'test_config.cfg')
        config.read(config_path)

        mysql_host = config.get('mysql', 'host')
        mysql_database = config.get('mysql', 'database')
        mysql_user = config.get('mysql', 'user')
        mysql_password = config.get('mysql', 'password')

        self.storage = UserStorage(mysql_host, mysql_user, mysql_password, mysql_database)

    def test_by_email_not_existing(self):

        user = self.storage.get_user_by_email('some_non_existing_email')
        self.assertUser(None, user)

    def test_insert(self):

        user = self.generate_user()
        self.storage.insert_user(user)
        read = self.storage.get_user_by_email(user.email)
        self.assertUser(user, read)

    def test_delete(self):

        user = self.generate_user()
        self.storage.insert_user(user)
        deleted = self.storage.delete_user(user.email)
        self.assertTrue(deleted)
        after_delete = self.storage.get_user_by_email(user.email)
        self.assertUser(None, after_delete)

    def test_by_token_not_existing(self):

        user = self.storage.get_user_by_email('token_not_existing')
        self.assertUser(None, user)

    def test_by_update_token_existing(self):

        user = self.generate_user()
        self.storage.insert_user(user)
        read = self.storage.get_user_by_token(user.update_token)
        self.assertUser(user, read)

    def test_by_domain_not_existing(self):

        user = self.storage.get_user_by_domain('domain_not_existing')
        self.assertUser(None, user)

    def test_by_update_token_existing(self):

        user = self.generate_user()
        self.storage.insert_user(user)
        read = self.storage.get_user_by_domain(user.user_domain)
        self.assertUser(user, read)

if __name__ == '__main__':
    unittest.run()
