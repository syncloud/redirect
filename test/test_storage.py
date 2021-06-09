import unittest

from helpers import generate_user, ModelsAssertionsMixin
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

    def test_user_add_different_session(self):
        user = generate_user()
        with self.create_storage() as storage:
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


if __name__ == '__main__':
    unittest.run()
