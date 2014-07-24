import redirect.rest
import unittest
import datetime

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

    def acquire_domain(self, email, password, user_domain):
        response = self.app.post('/domain/acquire', data=dict(user_domain=user_domain, email=email, password=password))
        domain_data = json.loads(response.data)
        update_token = domain_data['update_token']
        return update_token

class TestUser(TestFlask):

    def test_user_create_success(self):
        user_domain = create_token()
        email = user_domain+'@mail.com'
        response = self.app.post('/user/create', data={'email': email, 'password': 'pass123456'})
        self.assertEqual(200, response.status_code)
        self.assertFalse(self.smtp.empty())

    def test_user_get_success(self):
        email, password = self.create_active_user()

        response = self.app.get('/user/get', query_string={'email': email, 'password': password})
        self.assertEqual(200, response.status_code, response.data)

    def test_user_activate_success(self):
        user_domain = create_token()
        email = user_domain+'@mail.com'
        self.app.post('/user/create', data={'email': email, 'password': 'pass123456'})

        self.assertFalse(self.smtp.empty())
        token = self.get_token(self.smtp.emails()[0])

        activate_response = self.app.get('/user/activate', query_string={'token': token})
        self.assertEqual(200, activate_response.status_code)

    def test_get_user_data(self):
        email, password = self.create_active_user()

        user_domain = create_token()
        update_token = self.acquire_domain(email, password, user_domain)

        service_data = {'name': 'ownCloud', 'protocol': 'http', 'type': '_http._tcp', 'port': 10000, 'url': None}
        update_data = {'token': update_token, 'ip': '127.0.0.1', 'services': [service_data]}
        self.app.post('/domain/update', data=json.dumps(update_data))

        response = self.app.get('/user/get', query_string={'email': email, 'password': password})

        response_data = json.loads(response.data)
        user_data = response_data['data']

        # This is hack. We do not know last_update value - it is set by server.
        last_update = user_data["domains"][0]["last_update"]

        expected = {
            "active": True,
            "email": email,
            "domains": [{
                "user_domain": user_domain,
                "ip": "127.0.0.1",
                "last_update": last_update,
                "services": [{
                    "name": "ownCloud",
                    "protocol": "http",
                    "port": 10000,
                    "type": "_http._tcp",
                    "url": None
                }],
            }]
        }

        self.assertEquals(expected, user_data)

    def test_user_delete(self):
        email, password = self.create_active_user()

        user_domain_1 = create_token()
        update_token_1 = self.acquire_domain(email, password, user_domain_1)

        user_domain_2 = create_token()
        update_token_2 = self.acquire_domain(email, password, user_domain_2)

        response = self.app.post('/user/delete', data={'email': email, 'password': password})
        self.assertEquals(200, response.status_code)

        response = self.app.get('/domain/get', query_string={'token': update_token_1})
        self.assertEqual(400, response.status_code)

        response = self.app.get('/domain/get', query_string={'token': update_token_2})
        self.assertEqual(400, response.status_code)


class TestDomain(TestFlask):

    def assertDomain(self, expected, actual):
        for key, expected_value in expected.items():
            actual_value = actual[key]
            self.assertEquals(expected_value, actual_value)

    def check_domain(self, update_token, expected_data):
        response = self.app.get('/domain/get', query_string={'token': update_token})
        self.assertEqual(200, response.status_code)
        self.assertIsNotNone(response.data)
        response_data = json.loads(response.data)
        domain_data = response_data['data']
        self.assertDomain(expected_data, domain_data)
        return domain_data

    def test_acquire_new(self):
        email, password = self.create_active_user()

        user_domain = create_token()
        response = self.app.post('/domain/acquire', data=dict(user_domain=user_domain, email=email, password=password))

        self.assertEqual(200, response.status_code)
        domain_data = json.loads(response.data)

        update_token = domain_data['update_token']
        self.assertIsNotNone(update_token)

        self.check_domain(update_token, {'ip': None, 'user_domain': user_domain, 'services': []})

    def test_acquire_existing(self):
        user_domain = create_token()

        other_email, other_password = self.create_active_user()
        response = self.app.post('/domain/acquire', data=dict(user_domain=user_domain, email=other_email, password=other_password))
        domain_data = json.loads(response.data)
        update_token = domain_data['update_token']

        email, password = self.create_active_user()
        response = self.app.post('/domain/acquire', data=dict(user_domain=user_domain, email=email, password=password))

        self.assertEqual(409, response.status_code)

        self.check_domain(update_token, {'ip': None, 'user_domain': user_domain, 'services': []})

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

        self.check_domain(update_token2, {'ip': None, 'user_domain': user_domain, 'services': []})

    def test_domain_update_one_new_service(self):
        email, password = self.create_active_user()

        user_domain = create_token()
        response = self.app.post('/domain/acquire', data=dict(user_domain=user_domain, email=email, password=password))

        domain_data = json.loads(response.data)
        update_token = domain_data['update_token']

        service_data = {'name': 'ownCloud', 'protocol': 'http', 'type': '_http._tcp', 'port': 10000, 'url': None}
        update_data = {'token': update_token, 'ip': '127.0.0.1', 'services': [service_data]}

        response = self.app.post('/domain/update', data=json.dumps(update_data))
        self.assertEqual(200, response.status_code)

        self.check_domain(update_token, {'ip': '127.0.0.1', 'user_domain': user_domain, 'services': [service_data]})

    def test_domain_update_wrong_token(self):
        service_data = {'name': 'ownCloud', 'type': '_http._tcp', 'port': 10000, 'url': None}
        update_data = {'token': create_token(), 'ip': '127.0.0.1', 'services': [service_data]}

        response = self.app.post('/domain/update', data=json.dumps(update_data))
        self.assertEqual(400, response.status_code)

    def test_domain_update_missing_port(self):
        email, password = self.create_active_user()

        user_domain = create_token()
        response = self.app.post('/domain/acquire', data=dict(user_domain=user_domain, email=email, password=password))

        domain_data = json.loads(response.data)
        update_token = domain_data['update_token']

        service_data = {'name': 'ownCloud', 'type': '_http._tcp', 'url': None}
        update_data = {'token': update_token, 'ip': '127.0.0.1', 'services': [service_data]}

        response = self.app.post('/domain/update', data=json.dumps(update_data))
        self.assertEqual(400, response.status_code)

    def test_domain_update_two_new_services(self):
        email, password = self.create_active_user()

        user_domain = create_token()
        response = self.app.post('/domain/acquire', data=dict(user_domain=user_domain, email=email, password=password))

        domain_data = json.loads(response.data)
        update_token = domain_data['update_token']

        services_data = [{'name': 'ownCloud', 'protocol': 'http', 'type': '_http._tcp', 'port': 10001, 'url': None},
                         {'name': 'SSH', 'protocol': 'https', 'type': '_http._tcp', 'port': 10002, 'url': None}]
        update_data = {'token': update_token, 'ip': '127.0.0.1', 'services': services_data}

        response = self.app.post('/domain/update', data=json.dumps(update_data))
        self.assertEqual(200, response.status_code)

        self.check_domain(update_token, {'ip': '127.0.0.1', 'user_domain': user_domain, 'services': services_data})

    def test_domain_update_service_removed(self):
        email, password = self.create_active_user()

        user_domain = create_token()
        response = self.app.post('/domain/acquire', data=dict(user_domain=user_domain, email=email, password=password))

        domain_data = json.loads(response.data)
        update_token = domain_data['update_token']

        services_data = [{'name': 'ownCloud', 'protocol': 'http', 'type': '_http._tcp', 'port': 10001, 'url': None},
                         {'name': 'SSH', 'protocol': 'https', 'type': '_http._tcp', 'port': 10002, 'url': None}]
        update_data = {'token': update_token, 'ip': '127.0.0.1', 'services': services_data}

        self.app.post('/domain/update', data=json.dumps(update_data))

        services_data = [{'name': 'ownCloud', 'protocol': 'http', 'type': '_http._tcp', 'port': 10001, 'url': None}]
        update_data = {'token': update_token, 'ip': '127.0.0.1', 'services': services_data}

        response = self.app.post('/domain/update', data=json.dumps(update_data))

        self.assertEqual(200, response.status_code)

        self.check_domain(update_token, {'ip': '127.0.0.1', 'user_domain': user_domain, 'services': services_data})

    def test_domain_update_service_updated(self):
        email, password = self.create_active_user()

        user_domain = create_token()
        response = self.app.post('/domain/acquire', data=dict(user_domain=user_domain, email=email, password=password))

        domain_data = json.loads(response.data)
        update_token = domain_data['update_token']

        services_data = [{'name': 'ownCloud', 'protocol': 'http', 'type': '_http._tcp', 'port': 10001, 'url': None},
                         {'name': 'SSH', 'protocol': 'https', 'type': '_http._tcp', 'port': 10002, 'url': None}]
        update_data = {'token': update_token, 'ip': '127.0.0.1', 'services': services_data}

        self.app.post('/domain/update', data=json.dumps(update_data))

        services_data = [{'name': 'ownCloud', 'protocol': 'http', 'type': '_http._tcp', 'port': 10001, 'url': None},
                         {'name': 'SSH', 'protocol': 'https', 'type': '_http._tcp', 'port': 10003, 'url': None}]
        update_data = {'token': update_token, 'ip': '127.0.0.1', 'services': services_data}

        response = self.app.post('/domain/update', data=json.dumps(update_data))

        self.assertEqual(200, response.status_code)

        self.check_domain(update_token, {'ip': '127.0.0.1', 'user_domain': user_domain, 'services': services_data})

    def test_domain_update_ip_changed(self):
        email, password = self.create_active_user()

        user_domain = create_token()
        response = self.app.post('/domain/acquire', data=dict(user_domain=user_domain, email=email, password=password))

        domain_data = json.loads(response.data)
        update_token = domain_data['update_token']

        services_data = [{'name': 'ownCloud', 'type': '_http._tcp', 'port': 10001, 'url': None}]
        update_data = {'token': update_token, 'ip': '127.0.0.1', 'services': services_data}

        self.app.post('/domain/update', data=json.dumps(update_data))

        services_data = [{'name': 'ownCloud', 'protocol': 'http', 'type': '_http._tcp', 'port': 10001, 'url': None}]
        update_data = {'token': update_token, 'ip': '127.0.0.2', 'services': services_data}

        response = self.app.post('/domain/update', data=json.dumps(update_data))

        self.assertEqual(200, response.status_code)

        self.check_domain(update_token, {'ip': '127.0.0.2', 'user_domain': user_domain, 'services': services_data})