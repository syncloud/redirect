import redirect.rest
import redirect.www
import unittest
import datetime
import time
from syncloudlib.json import convertible

import json

from redirect.util import create_token
from fakesmtp import FakeSmtp
from urlparse import urlparse


class TestFlask(unittest.TestCase):

    def setUp(self):
        redirect.rest.app.config['TESTING'] = True
        redirect.www.app.config['TESTING'] = True
        self.app = redirect.rest.app.test_client()
        self.www = redirect.www.app.test_client()

        self.smtp = FakeSmtp('localhost', 2500)

    def tearDown(self):
        self.smtp.stop()




    def get_user(self, email, password):
        response = self.app.get('/user/get', query_string={'email': email, 'password': password})

        response_data = json.loads(response.data)
        user_data = response_data['data']

        return user_data


class TestDomain(TestFlask):

    def assertDomain(self, expected, actual):
        for key, expected_value in expected.items():
            actual_value = actual[key]
            self.assertEquals(expected_value, actual_value, 'Key "{}" has different values: {} != {}'.format(key, expected_value, actual_value))

    def get_domain(self, update_token):
        response = self.app.get('/domain/get', query_string={'token': update_token})
        self.assertEqual(200, response.status_code)
        self.assertIsNotNone(response.data)
        response_data = json.loads(response.data)
        return response_data['data']

    def check_domain(self, update_token, expected_data):
        domain_data = self.get_domain(update_token)
        self.assertDomain(expected_data, domain_data)
        return domain_data


class TestDomainAcquire(TestDomain):






class TestDomainLoose(TestDomain):

    def test_simple(self):
        email, password = self.create_active_user()

        user_domain = create_token()
        acquire_data = dict(
            user_domain=user_domain,
            device_mac_address='00:00:00:00:00:00',
            device_name='my-super-board',
            device_title='My Super Board',
            email=email,
            password=password)
        response = self.app.post('/domain/acquire', data=acquire_data)

        self.assertEqual(200, response.status_code)
        domain_data = json.loads(response.data)

        update_token = domain_data['update_token']

        drop_data = {
            'email': email,
            'password': password,
            'user_domain': user_domain
        }

        response = self.app.post('/domain/drop_device', data=drop_data)
        self.assertEqual(200, response.status_code)

        response = self.app.get('/domain/get', query_string={'token': update_token})
        self.assertEqual(400, response.status_code)


class TestDomainDelete(TestDomain):

    def test_simple(self):
        email, password = self.create_active_user()

        user_domain = create_token()
        update_token = self.acquire_domain(email, password, user_domain)

        delete_data = {'user_domain': user_domain, 'email': email, 'password': password}

        response = self.app.post('/domain/delete', data=json.dumps(delete_data))
        self.assertEqual(200, response.status_code)

        user_data = self.get_user(email, password)

        self.assertEquals(0, len(user_data['domains']))


class TestDomainUpdate(TestDomain):

    def test_domain_update_date(self):
        email, password = self.create_active_user()

        user_domain = create_token()

        update_token = self.acquire_domain(email, password, user_domain)

        update_data = {
            'token': update_token,
            'ip': '127.0.0.1',
            'web_protocol': 'http',
            'web_port': 10001,
            'web_local_port': 80
        }

        self.app.post('/domain/update', data=json.dumps(update_data))
        domain = self.get_domain(update_token)
        last_updated1 = domain['last_update']

        time.sleep(1)

        self.app.post('/domain/update', data=json.dumps(update_data))
        domain = self.get_domain(update_token)
        last_updated2 = domain['last_update']

        self.assertGreater(last_updated2, last_updated1)

    def test_domain_update_wrong_token(self):
        update_data = {'token': create_token(), 'ip': '127.0.0.1'}

        response = self.app.post('/domain/update', data=json.dumps(update_data))
        self.assertEqual(400, response.status_code)

    def test_domain_update_web_updated(self):
        email, password = self.create_active_user()

        user_domain = create_token()
        update_token = self.acquire_domain(email, password, user_domain)

        update_data = {
            'token': update_token,
            'ip': '127.0.0.1',
            'web_protocol': 'http',
            'web_port': 10001,
            'web_local_port': 80,
        }

        response = self.app.post('/domain/update', data=json.dumps(update_data))
        self.assertEqual(200, response.status_code)

        update_data = {
            'token': update_token,
            'ip': '127.0.0.1',
            'web_protocol': 'https',
            'web_port': 10002,
            'web_local_port': 443,
        }

        response = self.app.post('/domain/update', data=json.dumps(update_data))

        self.assertEqual(200, response.status_code)

        expected_data = {
            'ip': '127.0.0.1',
            'user_domain': user_domain,
            'web_protocol': 'https',
            'web_port': 10002,
            'web_local_port': 443,
        }

        self.check_domain(update_token, expected_data)

    def test_domain_update_ip_changed(self):
        email, password = self.create_active_user()

        user_domain = create_token()
        update_token = self.acquire_domain(email, password, user_domain)

        update_data = {
            'token': update_token,
            'ip': '127.0.0.1',
            'web_protocol': 'http',
            'web_port': 10001,
            'web_local_port': 80,
        }

        response = self.app.post('/domain/update', data=json.dumps(update_data))
        self.assertEqual(200, response.status_code)

        update_data = {
            'token': update_token,
            'ip': '127.0.0.2',
            'web_protocol': 'http',
            'web_port': 10001,
            'web_local_port': 80,
        }

        response = self.app.post('/domain/update', data=json.dumps(update_data))

        self.assertEqual(200, response.status_code)

        self.check_domain(update_token, {'ip': '127.0.0.2', 'user_domain': user_domain})

    def test_domain_update_platform_version(self):
        email, password = self.create_active_user()

        user_domain = create_token()
        update_token = self.acquire_domain(email, password, user_domain)

        update_data = {
            'token': update_token,
            'ip': '127.0.0.1',
            'platform_version': '366',
            'web_protocol': 'http',
            'web_port': 10001,
            'web_local_port': 80,
        }

        response = self.app.post('/domain/update', data=json.dumps(update_data))
        self.assertEqual(200, response.status_code)

        self.check_domain(update_token, {'platform_version': '366'})

    def test_local_ip_changed(self):
        email, password = self.create_active_user()

        user_domain = create_token()
        update_token = self.acquire_domain(email, password, user_domain)

        update_data = {
            'token': update_token,
            'ip': '127.0.0.1',
            'local_ip': '192.168.1.5',
            'web_protocol': 'http',
            'web_port': 10001,
            'web_local_port': 80,
        }

        response = self.app.post('/domain/update', data=json.dumps(update_data))
        self.assertEqual(200, response.status_code)

        update_data = {
            'token': update_token,
            'ip': '127.0.0.2',
            'local_ip': '192.168.1.6',
            'web_protocol': 'http',
            'web_port': 10001,
            'web_local_port': 80,
        }

        response = self.app.post('/domain/update', data=json.dumps(update_data))

        self.assertEqual(200, response.status_code)

        self.check_domain(update_token, {'ip': '127.0.0.2', 'local_ip': '192.168.1.6', 'user_domain': user_domain})

    def test_domain_update_server_side_client_ip(self):
        email, password = self.create_active_user()

        user_domain = create_token()
        update_token = self.acquire_domain(email, password, user_domain)

        update_data = {
            'token': update_token,
            'web_protocol': 'http',
            'web_port': 10001,
            'web_local_port': 80,
        }

        response = self.app.post('/domain/update', data=json.dumps(update_data), environ_base={'REMOTE_ADDR': '127.0.0.1'})
        self.assertEqual(200, response.status_code)

        expected_data = {
            'ip': '127.0.0.1',
            'user_domain': user_domain,
            'web_protocol': 'http',
            'web_port': 10001,
            'web_local_port': 80,
        }

        self.check_domain(update_token, expected_data)

    def test_domain_update_map_local_address(self):
        email, password = self.create_active_user()

        user_domain = create_token()
        update_token = self.acquire_domain(email, password, user_domain)

        update_data = {
            'token': update_token,
            'ip': '108.108.108.108',
            'local_ip': '192.168.1.2',
            'map_local_address': True,
            'web_protocol': 'http',
            'web_port': 10001,
            'web_local_port': 80
        }

        response = self.app.post('/domain/update', data=json.dumps(update_data))
        self.assertEqual(200, response.status_code)

        self.check_domain(update_token, {'ip': '108.108.108.108', 'local_ip': '192.168.1.2', 'map_local_address': True, 'user_domain': user_domain})
