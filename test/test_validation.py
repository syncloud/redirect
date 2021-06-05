from redirect.validation import Validator
import unittest

valid_params = {
    'user_domain': 'username',
    'email': 'valid@mail.com',
    'password': 'pass123456',
    'port': '80',
    'ip': '192.168.1.1'}


class TestValidation(unittest.TestCase):

    def assertEmailError(self, params):

        validator = Validator(params)
        value = validator.email()
        self.assertEqual(len(validator.errors), 1)
        return value

    def assertNewPasswordError(self, params):

        validator = Validator(params)
        value = validator.new_password()
        self.assertEqual(len(validator.errors), 1)
        return value

    def test_email_missing(self):

        params = {}
        self.assertEmailError(params)

    def test_email_invalid(self):

        params = {'email': 'invalid.email'}
        self.assertEmailError(params)

    def test_password_missing(self):

        params = {}
        self.assertNewPasswordError(params)

    def test_password_short(self):

        params = {'password': '123456'}
        self.assertNewPasswordError(params)

    def test_ip_missing(self):

        params = {}
        validator = Validator(params)
        ip = validator.ip()
        self.assertIsNone(ip)
        self.assertEquals(0, len(validator.errors))

    def test_ip_default(self):

        params = {}
        validator = Validator(params)
        ip = validator.ip('192.168.0.1')
        self.assertEquals(ip, '192.168.0.1')
        self.assertEquals(0, len(validator.errors))

    def test_ip_invalid(self):

        params = {'ip': '256.256.256.256'}
        validator = Validator(params)
        ip = validator.ip()
        self.assertEqual(len(validator.errors), 1)

    def test_errors_aggregated(self):

        params = {}
        validator = Validator(params)
        validator.user_domain()
        validator.password()
        self.assertEquals(2, len(validator.errors))