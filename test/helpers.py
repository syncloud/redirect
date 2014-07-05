import os
import ConfigParser
from redirect.db_helper import get_storage_creator
from redirect.util import create_token, hash
from redirect.models import User, Domain


def get_test_storage_creator():
    return get_storage_creator(get_test_config())


def get_test_config():
    config = ConfigParser.ConfigParser()
    config_path = os.path.join(os.path.dirname(__file__), 'test_config.cfg')
    config.read(config_path)
    return config


def generate_user():
    email = unicode(create_token() + '@mail.com')
    user = User(email, hash('pass1234'), False)
    user.set_activate_token(create_token())
    return user


def generate_domain():
    domain = create_token()
    update_token = create_token()
    domain = Domain(domain, None, update_token)
    return domain


class ModelsAssertionsMixin:

    def assertUser(self, expected, actual):
        if expected is None:
            self.assertIsNone(actual)
        if expected is not None:
            self.assertIsNotNone(actual)
        if expected is not None and actual is not None:
            self.assertEquals(expected.email, actual.email, 'Users should have the same email')
            self.assertEquals(expected.password_hash, actual.password_hash, 'Users should have the same password_hash')
            self.assertEquals(expected.active, actual.active, 'Users should have the same active')
            self.assertEquals(expected.activate_token(), actual.activate_token(), 'Users should have the same activate_token')

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
