import unittest

from helpers import generate_user, ModelsAssertionsMixin
from redirect.models import ActionType

from redirect.util import hash
from test.helpers import get_test_storage_creator


class TestStorageUser(ModelsAssertionsMixin, unittest.TestCase):

    def setUp(self):
        self.create_storage = get_test_storage_creator()

    def tearDown(self):
        with self.create_storage() as storage:
            storage.clear()

    def test_by_email_not_existing(self):
        with self.create_storage() as storage:
            user = storage.get_user_by_email(u'some_non_existing_email')
        self.assertUser(None, user)

    def test_user_add(self):
        with self.create_storage() as storage:
            user = generate_user()
            storage.add(user)
            read = storage.get_user_by_email(user.email)
            self.assertUser(user, read)

    def test_get_users_emails(self):
        user = generate_user()
        with self.create_storage() as storage:
            storage.add(user)
        with self.create_storage() as storage:
            read = storage.get_users_emails('SELECT email FROM user WHERE email="{0}"'.format(user.email))
            self.assertEquals(1, len(read))
            self.assertEquals(user.email, read[0])

    def test_user_add_different_session(self):
        user = generate_user()
        with self.create_storage() as storage:
            storage.add(user)
        with self.create_storage() as storage:
            read = storage.get_user_by_email(user.email)
            self.assertUser(user, read)

    def test_user_password_hash_fits_column(self):
        with self.create_storage() as storage:
            user = generate_user()
            user.password_hash = hash(user.password_hash)
            storage.add(user)
        with self.create_storage() as storage:
            read = storage.get_user_by_email(user.email)
            self.assertUser(user, read)

    def test_update_user(self):
        user = generate_user()
        user.active = False
        with self.create_storage() as storage:
            storage.add(user)

        with self.create_storage() as storage:
            update = storage.get_user_by_email(user.email)
            update.active = True

        with self.create_storage() as storage:
            read = storage.get_user_by_email(user.email)
            self.assertTrue(read.active)

  
    
    def test_iterate_one_user(self):
        user = generate_user()
        with self.create_storage() as storage:
            storage.add(user)
        with self.create_storage() as storage:
            users = storage.users_iterate()
        self.assertEquals(1, len(list(users)))

    def test_iterate_two_users(self):
        user1 = generate_user()
        user2 = generate_user()
        with self.create_storage() as storage:
            storage.add(user1, user2)
        with self.create_storage() as storage:
            users = storage.users_iterate()
        self.assertEquals(2, len(list(users)))

    def test_iterate_user_unsubscribed(self):
        user = generate_user()
        user.unsubscribed = True
        with self.create_storage() as storage:
            storage.add(user)
        with self.create_storage() as storage:
            users = storage.users_iterate()
        self.assertEquals(0, len(list(users)))


if __name__ == '__main__':
    unittest.run()
