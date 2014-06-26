import redirect.rest
import unittest

import json

from redirect.util import create_token
from fakesmtp import FakeSmtp
from urlparse import urlparse

class FlaskrTestCase(unittest.TestCase):

    def setUp(self):
        redirect.rest.app.config['TESTING'] = True
        self.app = redirect.rest.app.test_client()

        self.smtp = FakeSmtp('outbox')
        self.smtp.clear()

    def tearDown(self):
        pass

    def get_token(self, email):
        link_index = email.find('http://')
        link = email[link_index:].split(' ')[0].strip()
        parts = urlparse(link)
        token = parts.query.replace('token=', '')
        return token

    def test_user_create_success(self):
        user_domain = create_token()
        email = user_domain+'@mail.com'
        params = {'user_domain': user_domain, 'email': email, 'password': 'pass123456'}
        response = self.app.post('/user/create', data=params)
        self.assertEqual(200, response.status_code)
        self.assertFalse(self.smtp.empty())

    def test_user_activate_success(self):
        user_domain = create_token()
        email = user_domain+'@mail.com'
        create_params = {'user_domain': user_domain, 'email': email, 'password': 'pass123456'}
        self.app.post('/user/create', data=create_params)

        self.assertFalse(self.smtp.empty())
        token = self.get_token(self.smtp.emails()[0])

        params = {'token': token}
        activate_response = self.app.get('/user/activate', query_string=params)
        self.assertEqual(200, activate_response.status_code)

    def test_domain_update(self):
        user_domain = create_token()
        email = user_domain+'@mail.com'
        self.app.post('/user/create', data={'user_domain': user_domain, 'email': email, 'password': 'pass123456'})
        activate_token = self.get_token(self.smtp.emails()[0])
        self.app.get('/user/activate', query_string={'token': activate_token})

        token_response = self.app.get('/user/get', query_string={'email': email, 'password': 'pass123456'})
        self.assertEqual(200, token_response.status_code)
        user_data = json.loads(token_response.data)

        update_token = user_data['update_token']

        response = self.app.post('/domain/update', data={'token': update_token, 'ip': '127.0.0.1', 'port': '10001'})
        self.assertEqual(200, response.status_code)