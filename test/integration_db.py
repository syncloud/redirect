import unittest
import ConfigParser
import os
from redirect.db import Db

config = ConfigParser.ConfigParser()
config.read(os.path.dirname(__file__) + '/test.config.cfg')

token = "token123"
user_domain = "user1"
domain = "example.com"
user_email = config.get('mail', 'user_email')
mail_from = config.get('mail', 'mail_from')


class TestIntegrationDb(unittest.TestCase):
    def test_get_port_by_user_domain(self):

        db = Db('localhost', 'root', 'root', 'redirect')
        db.insert('test_123', 'email', 'password', 'token', 'ip', '80')
        db.activate('token')
        port = db.get_port_by_user_domain('test_123')

        url = 'domain:{}/owncloud'.format(port)

        db.delete_user('test_123', 'password')

        self.assertEquals(url, 'domain:80/owncloud')




