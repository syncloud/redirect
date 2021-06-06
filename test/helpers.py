import os

import ConfigParser

from redirect.db_helper import get_storage_creator
from redirect.models import User, ActionType
from redirect.util import create_token, hash


def get_test_storage_creator():
    return get_storage_creator(get_test_config())


def get_test_config():
    config = ConfigParser.ConfigParser()
    config_path = os.path.join(os.path.dirname(__file__), 'config.cfg')
    config.read(config_path)
    return config


def generate_user():
    email = unicode(create_token() + '@mail.com')
    user = User(email, hash('pass1234'), False)
    user.enable_action(ActionType.ACTIVATE)
    return user


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
            self.assertEquals(expected.token(ActionType.ACTIVATE), actual.token(ActionType.ACTIVATE), 'Users should have the same activation token')
            self.assertEquals(expected.token(ActionType.PASSWORD), actual.token(ActionType.PASSWORD), 'Users should have the same reset password token')

    def assertDomain(self, expected, actual):
        if expected is None:
            self.assertIsNone(actual)
        if expected is not None:
            self.assertIsNotNone(actual)
        if expected is not None and actual is not None:
            self.assertEquals(expected.user_domain, actual.user_domain, 'Domains should have the same user_domain')
            self.assertEquals(expected.ip, actual.ip, 'Domains should have the same ip')
            self.assertEquals(expected.local_ip, actual.local_ip, 'Domains should have the same local_ip')
            self.assertEquals(expected.update_token, actual.update_token, 'Domains should have the same update_token')
            self.assertEquals(expected.user_id, actual.user_id, 'Domains should have the same user_id')
            self.assertEquals(expected.device_mac_address, actual.device_mac_address, 'Domains should have the same device_mac_address')
            self.assertEquals(expected.device_name, actual.device_name, 'Domains should have the same device_name')
            self.assertEquals(expected.device_title, actual.device_title, 'Domains should have the same device_title')
