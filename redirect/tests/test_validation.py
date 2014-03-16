from .. validation import Validator
import unittest

valid_params = {
    'username': 'username',
    'email': 'valid@mail.com',
    'password': 'pass123456',
    'port': '80',
    'ip': '192.168.1.1'}


class TestValidation(unittest.TestCase):

    def test_username_missing(self):

        params = dict(valid_params)
        del params['username']
        errors, _, _, _, _, _ = Validator().validate_create(params, None)

        self.assertEqual(len(errors), 1)

    def test_username_invalid(self):

        params = dict(valid_params)
        params['username'] = 'user.name'
        errors, _, _, _, _, _ = Validator().validate_create(params, None)

        self.assertEqual(len(errors), 1)

    def test_username_short(self):

        params = dict(valid_params)
        params['username'] = 'use'
        errors, _, _, _, _, _ = Validator().validate_create(params, None)

        self.assertEqual(len(errors), 1)

    def test_username_long(self):

        params = dict(valid_params)
        params['username'] = '12345678901234567890123456789012345678901234567890_'
        errors, _, _, _, _, _ = Validator().validate_create(params, None)

        self.assertEqual(len(errors), 1)

    def test_email_missing(self):

        params = dict(valid_params)
        del params['email']
        errors, _, _, _, _, _ = Validator().validate_create(params, None)

        self.assertEqual(len(errors), 1)

    def test_email_invalid(self):

        params = dict(valid_params)
        params['email'] = 'invalid.email'
        errors, _, _, _, _, _ = Validator().validate_create(params, None)

        self.assertEqual(len(errors), 1)

    def test_password_missing(self):

        params = dict(valid_params)
        del params['password']
        errors, _, _, _, _, _ = Validator().validate_create(params, None)

        self.assertEqual(len(errors), 1)

    def test_password_short(self):

        params = dict(valid_params)
        params['password'] = '123456'
        errors, _, _, _, _, _ = Validator().validate_create(params, None)

        self.assertEqual(len(errors), 1)

    def test_port_missing(self):

        params = dict(valid_params)
        del params['port']
        errors, _, _, _, _, _ = Validator().validate_create(params, None)

        self.assertEqual(len(errors), 1)

    def test_port_short(self):

        params = dict(valid_params)
        params['port'] = '0'
        errors, _, _, _, _, _ = Validator().validate_create(params, None)

        self.assertEqual(len(errors), 1)

    def test_port_long(self):

        params = dict(valid_params)
        params['port'] = '65536'
        errors, _, _, _, _, _ = Validator().validate_create(params, None)

        self.assertEqual(len(errors), 1)

    def test_ip_missing(self):

        params = dict(valid_params)
        del params['ip']
        errors, _, _, _, _, ip = Validator().validate_create(params, '192.192.192.192')

        self.assertEqual(len(errors), 0)
        self.assertEqual(ip, '192.192.192.192')

    def test_ip_invalid(self):

        params = dict(valid_params)
        params['ip'] = '256.256.256.256'
        errors, _, _, _, _, _ = Validator().validate_create(params, None)

        self.assertEqual(len(errors), 1)

    def test_all_valid(self):

        errors, username, email, password, port, ip = Validator().validate_create(valid_params, None)

        self.assertEqual(len(errors), 0)
        self.assertEqual(username, valid_params['username'])
        self.assertEqual(email, valid_params['email'])
        self.assertEqual(password, valid_params['password'])
        self.assertEqual(port, valid_params['port'])
        self.assertEqual(ip, valid_params['ip'])

    def test_delete_username_missing(self):

        params = {'password': 'pass123'}

        errors, username, password = Validator().validate_delete(params)

        self.assertEqual(len(errors), 1)
        self.assertEqual(password, 'pass123')
        self.assertFalse(username)

    def test_delete_password_missing(self):

        params = {'username': 'user'}

        errors, username, password = Validator().validate_delete(params)

        self.assertEqual(len(errors), 1)
        self.assertEqual(username, 'user')
        self.assertFalse(password)
