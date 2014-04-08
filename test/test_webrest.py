import unittest
import requests
import uuid
import json
from urlparse import urljoin
from fakesmtp import FakeSmtp
from urlparse import urlparse

class TestWebRest(unittest.TestCase):

    base_url = r'http://127.0.0.1:5000'

    def post(self, url, params):
        return requests.post(urljoin(self.base_url, url), data=params)

    def get(self, url, parameters):
        return requests.get(urljoin(self.base_url, url), params=parameters)

    def get_token(self, email):
        link_index = email.find('http://')
        link = email[link_index:].split(' ')[0].strip()
        parts = urlparse(link)
        token = parts.query.replace('token=', '')
        return token

    def setUp(self):
        self.smtp = FakeSmtp('outbox')
        self.smtp.clear()

    def test_user_create_success(self):
        user_domain = uuid.uuid4().hex
        email = user_domain+'@mail.com'
        params = {'user_domain': user_domain, 'email': email, 'password': 'pass123456'}
        response = self.post('/user/create', params)
        self.assertTrue(response.ok, 'Response was: '+str(response))
        self.assertEqual(200, response.status_code)
        self.assertFalse(self.smtp.empty())

    def test_user_activate_success(self):
        user_domain = uuid.uuid4().hex
        email = user_domain+'@mail.com'
        create_params = {'user_domain': user_domain, 'email': email, 'password': 'pass123456'}
        self.post('/user/create', create_params)

        self.assertFalse(self.smtp.empty())
        token = self.get_token(self.smtp.emails()[0])

        response = {'token': token}
        activate_response = self.get('/user/activate', response)
        self.assertTrue(activate_response.ok, 'Response was: '+str(activate_response))
        self.assertEqual(200, activate_response.status_code)

    def test_domain_update(self):
        user_domain = uuid.uuid4().hex
        email = user_domain+'@mail.com'
        self.post('/user/create', {'user_domain': user_domain, 'email': email, 'password': 'pass123456'})
        activate_token = self.get_token(self.smtp.emails()[0])
        self.get('/user/activate', {'token': activate_token})

        token_response = self.post('/domain/token', {'email': email, 'password': 'pass123456'})
        self.assertTrue(token_response.ok, 'Response was: '+str(token_response))
        self.assertEqual(200, token_response.status_code)
        token_data = json.loads(token_response.content)

        update_token = token_data['token']

        response = self.post('/domain/update', {'token': update_token, 'ip': '127.0.0.1', 'port': '10001'})
        self.assertTrue(response.ok, 'Response was: '+str(response))
        self.assertEqual(200, response.status_code)


if __name__ == '__main__':
    unittest.run()
