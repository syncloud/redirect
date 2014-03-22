import unittest
from mock import MagicMock
from ..accountmanager import AccountManager


class TestAccountManager(unittest.TestCase):

    def setUp(self):
        self.validator = MagicMock()
        self.mail = MagicMock()
        self.db = MagicMock()
        self.dns = MagicMock()
        self.request = MagicMock()

    def test_redirect_url_not_registered(self):

        self.db.get_port_by_username = MagicMock(return_value=None)
        self.db.connect = MagicMock(return_value=self.db)
        manager = AccountManager(self.validator, self.db, self.dns, "example.com", False, self.mail)

        self.request.url = 'http://user.example.com'
        url = manager.redirect_url(self.request, 'example.org')

        self.assertEquals(url, 'http://example.org')

    def test_redirect_url_registered(self):

        self.db.get_port_by_username = MagicMock(return_value=80)
        self.db.close = MagicMock()
        self.db.connect = MagicMock(return_value=self.db)

        manager = AccountManager(self.validator, self.db, self.dns, "example.com", False, self.mail)

        self.request.url = 'http://user.example.com'
        url = manager.redirect_url(self.request, 'example.org')

        self.assertEquals(url, 'http://device.user.example.com:80/owncloud')

    def test_request_account_invalid_input(self):

        self.validator.validate_create = MagicMock(return_value=('errors', None, None, None, None, None))

        manager = AccountManager(self.validator, self.db, self.dns, "example.com", False, self.mail)

        manager.request_account(self.request)

        self.assertFalse(self.db.connect.called)

    def test_request_account_existing_user(self):

        self.validator.validate_create = MagicMock(return_value=('', None, None, None, None, None))

        self.db.close = MagicMock()
        self.db.connect = MagicMock(return_value=self.db)
        self.db.exists = MagicMock(return_value=True)
        manager = AccountManager(self.validator, self.db, self.dns, "example.com", False, self.mail)

        (text, code, _) = manager.request_account(self.request)

        self.assertFalse(self.db.insert.called)
        self.assertEquals(code, 409)

    def test_request_account_new_user_by_header(self):

        self.validator.validate_create = MagicMock(return_value=('', None, None, None, None, None))

        self.db.close = MagicMock()
        self.db.connect = MagicMock(return_value=self.db)
        self.db.exists = MagicMock(return_value=False)
        manager = AccountManager(self.validator, self.db, self.dns, "example.com", False, self.mail)

        (text, code, headers) = manager.request_account(self.request)

        self.assertTrue(self.db.insert.called)
        self.assertFalse(self.dns.create_records.called)
        self.assertFalse(self.mail.send.called)
        self.assertTrue(headers)
        self.assertEquals(code, 200)

    def test_request_account_new_user_by_mail(self):

        self.validator.validate_create = MagicMock(return_value=('', None, None, None, None, None))

        self.db.close = MagicMock()
        self.db.connect = MagicMock(return_value=self.db)
        self.db.exists = MagicMock(return_value=False)
        manager = AccountManager(self.validator, self.db, self.dns, "example.com", True, self.mail)

        (text, code, headers) = manager.request_account(self.request)

        self.assertTrue(self.db.insert.called)
        self.assertFalse(self.dns.create_records.called)
        self.assertTrue(self.mail.send.called)
        self.assertFalse(headers)
        self.assertEquals(code, 200)

    def test_request_account_new_user_db_exception(self):

        self.validator.validate_create = MagicMock(return_value=('', None, None, None, None, None))

        self.db.close = MagicMock()
        self.db.connect = MagicMock(return_value=self.db)
        self.db.exists = MagicMock(return_value=False)
        self.db.insert = MagicMock(side_effect=Exception)
        manager = AccountManager(self.validator, self.db, self.dns, "example.com", False, self.mail)

        (text, code, _) = manager.request_account(self.request)

        self.assertFalse(self.dns.create_records.called)
        self.assertEquals(code, 500)

    def test_activate_no_token(self):

        self.validator.validate_token = MagicMock(return_value=(['no token'], None))
        manager = AccountManager(self.validator, self.db, self.dns, "example.com", False, self.mail)

        (text, code) = manager.activate(self.request)

        self.assertFalse(self.dns.update_records.called)
        self.assertEquals(code, 400)

    def test_activate_invalid_token(self):

        self.validator.validate_token = MagicMock(return_value=([], '1'))
        self.db.activate = MagicMock(return_value=False)
        manager = AccountManager(self.validator, self.db, self.dns, "example.com", False, self.mail)

        (text, code) = manager.activate(self.request)

        self.assertFalse(self.dns.update_records.called)
        self.assertEquals(code, 400)

    def test_activate_exception(self):

        self.validator.validate_token = MagicMock(return_value=([], '1'))
        self.db.activate = MagicMock(side_effect=Exception)
        manager = AccountManager(self.validator, self.db, self.dns, "example.com", False, self.mail)

        (text, code) = manager.activate(self.request)

        self.assertFalse(self.dns.update_records.called)
        self.assertEquals(code, 500)

    def test_activate_success(self):

        self.validator.validate_token = MagicMock(return_value=([], '1'))
        self.db.activate = MagicMock(return_value=True)
        self.db.get_user_info_by_token = MagicMock(return_value=('user', 'ip', 'port'))
        manager = AccountManager(self.validator, self.db, self.dns, "example.com", False, self.mail)

        (text, code) = manager.activate(self.request)

        self.assertEquals(code, 200)
        self.assertTrue(self.dns.create_records.called)

    def test_update_invalid_input(self):

        self.validator.validate_update = MagicMock(return_value=(['invalid input'], '1', 'ip', 'port'))

        manager = AccountManager(self.validator, self.db, self.dns, "example.com", False, self.mail)

        (text, code) = manager.update(self.request)

        self.assertFalse(self.dns.update_records.called)
        self.assertEquals(code, 400)

    def test_update_invalid_token(self):

        self.validator.validate_update = MagicMock(return_value=([], '1', 'ip', 'port'))

        self.db.existing_token = MagicMock(return_value=False)
        manager = AccountManager(self.validator, self.db, self.dns, "example.com", False, self.mail)

        (text, code) = manager.update(self.request)

        self.assertFalse(self.dns.update_records.called)
        self.assertEquals(code, 400)

    def test_update_not_modified(self):

        self.validator.validate_update = MagicMock(return_value=([], '1', 'ip', 'port'))

        self.db.existing_token = MagicMock(return_value=True)
        self.db.get_user_info_by_token = MagicMock(return_value=('user', 'ip', 'port'))

        manager = AccountManager(self.validator, self.db, self.dns, "example.com", False, self.mail)

        (text, code) = manager.update(self.request)

        self.assertFalse(self.dns.update_records.called)
        self.assertEquals(code, 304)

    def test_update_exception(self):

        self.validator.validate_update = MagicMock(return_value=([], '1', 'ip', 'port'))

        self.db.existing_token = MagicMock(side_effect=Exception)

        manager = AccountManager(self.validator, self.db, self.dns, "example.com", False, self.mail)

        (text, code) = manager.update(self.request)

        self.assertFalse(self.dns.update_records.called)
        self.assertEquals(code, 500)

    def test_update_success(self):

        self.validator.validate_update = MagicMock(return_value=([], '1', 'new_ip', 'new_port'))

        self.db.existing_token = MagicMock(return_value=True)
        self.db.get_user_info_by_token = MagicMock(return_value=('user', 'ip', 'port'))

        manager = AccountManager(self.validator, self.db, self.dns, "example.com", False, self.mail)

        (text, code) = manager.update(self.request)

        self.assertTrue(self.dns.update_records.called)
        self.assertEquals(code, 200)

    def test_delete_invalid_input(self):

        self.validator.validate_delete = MagicMock(return_value=(['invalid input'], 'user', 'pass'))

        manager = AccountManager(self.validator, self.db, self.dns, "example.com", False, self.mail)

        (text, code) = manager.delete(self.request)

        self.assertFalse(self.db.delete_user.called)
        self.assertFalse(self.dns.delete_records.called)
        self.assertEquals(code, 400)

    def test_invalid_user(self):

        self.validator.validate_delete = MagicMock(return_value=([], 'user', 'pass'))

        self.db.valid_user = MagicMock(return_value=False)

        manager = AccountManager(self.validator, self.db, self.dns, "example.com", False, self.mail)

        (text, code) = manager.delete(self.request)

        self.assertFalse(self.db.delete_user.called)
        self.assertFalse(self.dns.delete_records.called)
        self.assertEquals(code, 400)

    def test_delete_exception(self):

        self.validator.validate_delete = MagicMock(return_value=([], 'user', 'pass'))

        self.db.valid_user = MagicMock(return_value=True)
        self.db.delete_user = MagicMock(side_effect=Exception)

        manager = AccountManager(self.validator, self.db, self.dns, "example.com", False, self.mail)

        (text, code) = manager.delete(self.request)

        # self.assertFalse(self.db.delete_user.called)
        self.assertFalse(self.dns.delete_records.called)
        self.assertEquals(code, 500)

    def test_delete_success(self):

        self.validator.validate_delete = MagicMock(return_value=([], 'user', 'pass'))

        self.db.valid_user = MagicMock(return_value=True)
        self.db.get_user_info_by_password = MagicMock(return_value=('user', 'ip', 80))
        manager = AccountManager(self.validator, self.db, self.dns, "example.com", False, self.mail)

        (text, code) = manager.delete(self.request)

        self.assertTrue(self.db.delete_user.called)
        self.assertTrue(self.dns.delete_records.called)
        self.assertEquals(code, 200)