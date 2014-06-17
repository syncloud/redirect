import unittest

from helpers import mysql_spec_test, generate_user, generate_domain

from redirect.util import hash

from redirect.session_alchemy import get_session_maker, SessionContextFactory
from redirect.storage import Storage

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
            self.assertEquals(expected.activate_token, actual.activate_token, 'Users should have the same activate_token')

    def assertDomain(self, expected, actual):
        if expected is None:
            self.assertIsNone(actual)
        if expected is not None:
            self.assertIsNotNone(actual)
        if expected is not None and actual is not None:
            self.assertEquals(expected.user_domain, actual.user_domain, 'Users should have the same user_domain')
            self.assertEquals(expected.ip, actual.ip, 'Users should have the same ip')
            self.assertEquals(expected.update_token, actual.update_token, 'Users should have the same update_token')
            self.assertEquals(expected.user_id, actual.user_id, 'Users should have the same user_id')


    def setUp(self):
        spec = mysql_spec_test()
        maker = get_session_maker(spec)
        self.factory = SessionContextFactory(maker)

    def tearDown(self):
        with self.factory() as session:
            Storage(session).clear()

    def test_by_email_not_existing(self):
        with self.factory() as session:
            user = Storage(session).get_user_by_email(u'some_non_existing_email')
        self.assertUser(None, user)

    def test_user_add(self):
        with self.factory() as session:
            storage = Storage(session)
            user = generate_user()
            storage.add(user)
            read = storage.get_user_by_email(user.email)
            self.assertUser(user, read)

    def test_user_add_different_session(self):
        user = generate_user()
        with self.factory() as session:
            Storage(session).add(user)
        with self.factory() as session:
            read = Storage(session).get_user_by_email(user.email)
            self.assertUser(user, read)

    def test_user_delete(self):
        user = generate_user()
        with self.factory() as session:
            Storage(session).add(user)
        with self.factory() as session:
            storage = Storage(session)
            deleted = storage.delete_user(user.email)
            self.assertTrue(deleted)
            after_delete = storage.get_user_by_email(user.email)
            self.assertUser(None, after_delete)

    def test_by_activate_token_not_existing(self):
        with self.factory() as session:
            user = Storage(session).get_user_by_activate_token(u'token_not_existing')
            self.assertUser(None, user)

    def test_by_activate_token_existing(self):
        with self.factory() as session:
            storage = Storage(session)
            user = generate_user()
            storage.add(user)
            read = storage.get_user_by_activate_token(user.activate_token)
            self.assertUser(user, read)

    def test_user_password_hash_fits_column(self):
        with self.factory() as session:
            storage = Storage(session)
            user = generate_user()
            user.password_hash = hash(user.password_hash)
            storage.add(user)
        with self.factory() as session:
            read = Storage(session).get_user_by_email(user.email)
            self.assertUser(user, read)

    def test_update_user(self):
        user = generate_user()
        user.active = False
        with self.factory() as session:
            Storage(session).add(user)

        with self.factory() as session:
            storage = Storage(session)
            update = storage.get_user_by_email(user.email)
            update.active = True

        with self.factory() as session:
            read = Storage(session).get_user_by_email(user.email)
            self.assertTrue(read.active)

    def test_domain_by_update_token_not_existing(self):
        with self.factory() as session:
            domain = Storage(session).get_domain_by_update_token(u'token_not_existing')
            self.assertDomain(None, domain)

    def test_domain_by_update_token_existing(self):
        user = generate_user()
        domain = generate_domain()
        domain.user = user
        with self.factory() as session:
            storage = Storage(session)
            storage.add(user)
            storage.add(domain)
        with self.factory() as session:
            read = Storage(session).get_domain_by_update_token(domain.update_token)
        self.assertDomain(domain, read)
        self.assertUser(user, read.user)

    def test_domain_by_name_not_existing(self):
        with self.factory() as session:
            domain = Storage(session).get_domain_by_name(u'domain_not_existing')
            self.assertUser(None, domain)

    def test_domain_by_name_existing(self):
        user = generate_user()
        domain = generate_domain()
        domain.user = user
        with self.factory() as session:
            storage = Storage(session)
            storage.add(user)
            storage.add(domain)
        with self.factory() as session:
            read = Storage(session).get_domain_by_name(domain.user_domain)
        self.assertDomain(domain, read)
        self.assertUser(user, read.user)

    def test_domain_by_name_existing(self):
        user = generate_user()
        domain = generate_domain()
        domain.user = user
        with self.factory() as session:
            storage = Storage(session)
            storage.add(user)
            storage.add(domain)
        with self.factory() as session:
            read = Storage(session).get_domain_by_name(domain.user_domain)
        self.assertDomain(domain, read)
        self.assertUser(user, read.user)

    def test_clear(self):
        user = generate_user()
        domain = generate_domain()
        domain.user = user
        with self.factory() as session:
            storage = Storage(session)
            storage.add(user)
            storage.add(domain)

        with self.factory() as session:
            Storage(session).clear()

        with self.factory() as session:
            storage = Storage(session)
            read_domain = storage.get_domain_by_name(domain.user_domain)
            read_user = storage.get_user_by_email(user.email)

        self.assertUser(None, read_user)
        self.assertDomain(None, read_domain)

if __name__ == '__main__':
    unittest.run()
