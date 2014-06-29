import redirect.rest
import unittest

import json

from redirect.util import create_token
from fakesmtp import FakeSmtp
from urlparse import urlparse

class TestFlask(unittest.TestCase):

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

    def create_active_user(self):
        self.smtp.clear()
        email = create_token()+'@mail.com'
        password = 'pass123456'
        self.app.post('/user/create', data={'email': email, 'password': password})
        activate_token = self.get_token(self.smtp.emails()[0])
        self.app.get('/user/activate', query_string={'token': activate_token})
        return email, password


class TestUser(TestFlask):

    def test_user_create_success(self):
        user_domain = create_token()
        email = user_domain+'@mail.com'
        response = self.app.post('/user/create', data={'email': email, 'password': 'pass123456'})
        self.assertEqual(200, response.status_code)
        self.assertFalse(self.smtp.empty())

    def test_user_activate_success(self):
        user_domain = create_token()
        email = user_domain+'@mail.com'
        self.app.post('/user/create', data={'email': email, 'password': 'pass123456'})

        self.assertFalse(self.smtp.empty())
        token = self.get_token(self.smtp.emails()[0])

        activate_response = self.app.get('/user/activate', query_string={'token': token})
        self.assertEqual(200, activate_response.status_code)


class TestDomain(TestFlask):

    def test_acquire_new(self):
        email, password = self.create_active_user()

        user_domain = create_token()
        response = self.app.post('/domain/acquire', data=dict(user_domain=user_domain, email=email, password=password))

        self.assertEqual(200, response.status_code)
        domain_data = json.loads(response.data)

        update_token = domain_data['update_token']
        self.assertIsNotNone(update_token)

    def test_acquire_existing(self):
        user_domain = create_token()

        other_email, other_password = self.create_active_user()
        self.app.post('/domain/acquire', data=dict(user_domain=user_domain, email=other_email, password=other_password))

        email, password = self.create_active_user()
        response = self.app.post('/domain/acquire', data=dict(user_domain=user_domain, email=email, password=password))

        self.assertEqual(409, response.status_code)

    def test_acquire_twice(self):
        email, password = self.create_active_user()

        user_domain = create_token()
        response = self.app.post('/domain/acquire', data=dict(user_domain=user_domain, email=email, password=password))
        domain_data = json.loads(response.data)
        update_token1 = domain_data['update_token']

        response = self.app.post('/domain/acquire', data=dict(user_domain=user_domain, email=email, password=password))
        self.assertEqual(200, response.status_code)
        domain_data = json.loads(response.data)
        update_token2 = domain_data['update_token']

        self.assertNotEquals(update_token1, update_token2)