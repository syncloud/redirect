from .. validation import Validator, Validation
import unittest

valid_params = {
    'username': 'username',
    'email': 'valid@mail.com',
    'password': 'pass123456',
    'port': '80',
    'ip': '192.168.1.1'}


class TestValidation(unittest.TestCase):

    def assertUsernameError(self, params):

        validator = Validator(params)
        value = validator.new_username()
        self.assertEqual(len(validator.errors), 1)
        return value

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

    def assertPortError(self, params):

        validator = Validator(params)
        value = validator.port()
        self.assertEqual(len(validator.errors), 1)
        return value

    def test_new_username_missing(self):

        params = {}
        username = self.assertUsernameError(params)
        self.assertIsNone(username)

    def test_new_username_invalid(self):

        params = {'username': 'user.name'}
        self.assertUsernameError(params)

    def test_username_short(self):

        params = {'username': 'use'}
        self.assertUsernameError(params)

    def test_username_long(self):

        params = {'username': '12345678901234567890123456789012345678901234567890_'}
        self.assertUsernameError(params)

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

    def test_ip_invalid(self):

        params = {'ip': '256.256.256.256'}
        validator = Validator(params)
        ip = validator.ip()
        self.assertEqual(len(validator.errors), 1)

    def test_port_missing(self):

        params = {}
        self.assertPortError(params)

    def test_port_small(self):

        params = {'port': '0'}
        self.assertPortError(params)

    def test_port_big(self):

        params = {'port': '65536'}
        self.assertPortError(params)

    def test_errors_aggregated(self):

        params = {}
        validator = Validator(params)
        validator.username()
        validator.password()
        self.assertEquals(2, len(validator.errors))

    def test_all_valid(self):

        errors, username, email, password, port, ip = Validation().validate_create(valid_params)

        self.assertEqual(len(errors), 0)
        self.assertEqual(username, 'username')
        self.assertEqual(email, 'valid@mail.com')
        self.assertEqual(password, 'pass123456')
        self.assertEqual(port, 80)
        self.assertEqual(ip, '192.168.1.1')

    def test_delete_username_missing(self):

        params = {'password': 'pass123'}

        errors, username, password = Validation().validate_credentials(params)

        self.assertEqual(len(errors), 1)
        self.assertEqual(password, 'pass123')
        self.assertFalse(username)

    def test_delete_password_missing(self):

        params = {'username': 'user'}

        errors, username, password = Validation().validate_credentials(params)

        self.assertEqual(len(errors), 1)
        self.assertEqual(username, 'user')
        self.assertFalse(password)
