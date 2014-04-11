import unittest
from redirect.mail import Mail
import ConfigParser
import os

config = ConfigParser.ConfigParser()
config.read(os.path.dirname(__file__) + '/test.config.cfg')

token = "token123"
user_domain = "user1"
domain = "example.com"
user_email = config.get('mail', 'user_email')
mail_from = config.get('mail', 'mail_from')


class TestIntegrationMail(unittest.TestCase):

    def test_send(self):

        mail = Mail(domain, mail_from)
        full_domain = '{0}.{1}'.format(user_domain, domain)
        mail.send_activate(full_domain, user_email, token)
