import redirect.rest
import unittest
import datetime
import time

import json

from redirect.util import create_token
from fakesmtp import FakeSmtp
from urlparse import urlparse

class TestFlask(unittest.TestCase):

    def setUp(self):
        redirect.rest.app.config['TESTING'] = True
        self.app = redirect.rest.app.test_client()

        self.smtp = FakeSmtp('outbox', 'localhost', 2500)

    def tearDown(self):
        self.smtp.stop()

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
        self.smtp.clear()
        return email, password

    def acquire_domain(self, email, password, user_domain):
        self.maxDiff = None
        acquire_data = {
            'user_domain': user_domain,
            'email': email,
            'password': password,
            'device_mac_address': '00:00:00:00:00:00',
            'device_name': 'some-device',
            'device_title': 'Some Device',
        }
        response = self.app.post('/domain/acquire', data=acquire_data)
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
            'active': True,
            'email': email,
            'domains': [{
                'user_domain': user_domain,
                'ip': '127.0.0.1',
                'local_ip': None,
                'device_mac_address': '00:00:00:00:00:00',
                'device_name': 'some-device',
                'device_title': 'Some Device',
                'last_update': last_update,
                'services': [{
                    'name': 'ownCloud',
                    'protocol': 'http',
                    'port': 10000,
                    'local_port': None,
                    'type': '_http._tcp',
                    'url': None
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


class TestUserPassword(TestFlask):

    def test_user_reset_password_sent_mail(self):
        email, password = self.create_active_user()

        response = self.app.post('/user/reset_password', data={'email': email})
        self.assertEqual(200, response.status_code)

        self.assertFalse(self.smtp.empty(), msg='Server should send email with link to reset password')
        token = self.get_token(self.smtp.emails()[0])

        self.assertIsNotNone(token)

    def test_user_reset_password_set_new(self):
        email, password = self.create_active_user()

        self.app.post('/user/reset_password', data={'email': email})
        token = self.get_token(self.smtp.emails()[0])

        self.smtp.clear()

        new_password = 'new_password'
        response = self.app.post('/user/set_password', data={'token': token, 'password': new_password})
        self.assertEqual(200, response.status_code, response.data)

        self.assertFalse(self.smtp.empty(), msg='Server should send email when setting new password')

        response = self.app.get('/user/get', query_string={'email': email, 'password': new_password})
        self.assertEqual(200, response.status_code, response.data)

    def test_user_reset_password_set_with_old_token(self):
        email, password = self.create_active_user()

        self.app.post('/user/reset_password', data={'email': email})
        token_old = self.get_token(self.smtp.emails()[0])

        self.smtp.clear()

        self.app.post('/user/reset_password', data={'email': email})
        token = self.get_token(self.smtp.emails()[0])

        self.smtp.clear()

        new_password = 'new_password'
        response = self.app.post('/user/set_password', data={'token': token_old, 'password': new_password})
        self.assertEqual(403, response.status_code, response.data)

    def test_user_reset_password_set_twice(self):
        email, password = self.create_active_user()

        self.app.post('/user/reset_password', data={'email': email})
        token = self.get_token(self.smtp.emails()[0])

        self.smtp.clear()

        new_password = 'new_password'
        response = self.app.post('/user/set_password', data={'token': token, 'password': new_password})
        self.assertEqual(200, response.status_code, response.data)

        new_password = 'new_password2'
        response = self.app.post('/user/set_password', data={'token': token, 'password': new_password})
        self.assertEqual(403, response.status_code, response.data)


class TestDomain(TestFlask):

    def assertDomain(self, expected, actual):
        for key, expected_value in expected.items():
            actual_value = actual[key]
            self.assertEquals(expected_value, actual_value)

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

    def test_new(self):
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
        self.assertIsNotNone(update_token)

        self.check_domain(update_token, {
            'ip': None,
            'user_domain': user_domain,
            'device_mac_address': '00:00:00:00:00:00',
            'device_name': 'my-super-board',
            'device_title': 'My Super Board',
            'services': []})

    def test_existing(self):
        user_domain = create_token()

        other_email, other_password = self.create_active_user()
        acquire_data = dict(
            user_domain=user_domain,
            device_mac_address='00:00:00:00:00:00',
            device_name='my-super-board',
            device_title='My Super Board',
            email=other_email,
            password=other_password)
        response = self.app.post('/domain/acquire', data=acquire_data)
        domain_data = json.loads(response.data)
        update_token = domain_data['update_token']

        email, password = self.create_active_user()
        acquire_data = dict(
            user_domain=user_domain,
            device_mac_address='00:00:00:00:00:11',
            device_name='other-board',
            device_title='Other Board',
            email=email,
            password=password)
        response = self.app.post('/domain/acquire', data=acquire_data)

        self.assertEqual(409, response.status_code)

        self.check_domain(update_token, {
            'ip': None,
            'user_domain': user_domain,
            'device_mac_address': '00:00:00:00:00:00',
            'device_name': 'my-super-board',
            'device_title': 'My Super Board',
            'services': []})

    def test_twice(self):
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
        domain_data = json.loads(response.data)
        update_token1 = domain_data['update_token']

        acquire_data = dict(
            user_domain=user_domain,
            device_mac_address='00:00:00:00:00:11',
            device_name='my-super-board-2',
            device_title='My Super Board 2',
            email=email,
            password=password)
        response = self.app.post('/domain/acquire', data=acquire_data)
        self.assertEqual(200, response.status_code)
        domain_data = json.loads(response.data)
        update_token2 = domain_data['update_token']

        self.assertNotEquals(update_token1, update_token2)

        self.check_domain(update_token2, {
            'ip': None,
            'user_domain': user_domain,
            'device_mac_address': '00:00:00:00:00:11',
            'device_name': 'my-super-board-2',
            'device_title': 'My Super Board 2',
            'services': []})

    def test_wrong_mac_address_format(self):
        email, password = self.create_active_user()

        user_domain = create_token()
        acquire_data = dict(
            user_domain=user_domain,
            device_mac_address='wrong_format',
            device_name='my-super-board',
            device_title='My Super Board',
            email=email,
            password=password)
        response = self.app.post('/domain/acquire', data=acquire_data)

        self.assertEqual(400, response.status_code)

class TestDomainUpdate(TestDomain):

    def test_domain_update_one_new_service(self):
        email, password = self.create_active_user()

        user_domain = create_token()
        update_token = self.acquire_domain(email, password, user_domain)

        service_data = {'name': 'ownCloud', 'protocol': 'http', 'type': '_http._tcp', 'port': 10000, 'local_port': None, 'url': None}
        update_data = {'token': update_token, 'ip': '127.0.0.1', 'services': [service_data]}

        response = self.app.post('/domain/update', data=json.dumps(update_data))
        self.assertEqual(200, response.status_code)

        self.check_domain(update_token, {'ip': '127.0.0.1', 'user_domain': user_domain, 'services': [service_data]})

    def test_domain_update_date(self):
        email, password = self.create_active_user()

        user_domain = create_token()

        update_token = self.acquire_domain(email, password, user_domain)

        update_data = {'token': update_token, 'ip': '127.0.0.1', 'services': []}

        self.app.post('/domain/update', data=json.dumps(update_data))
        domain = self.get_domain(update_token)
        last_updated1 = domain['last_update']

        time.sleep(1)

        self.app.post('/domain/update', data=json.dumps(update_data))
        domain = self.get_domain(update_token)
        last_updated2 = domain['last_update']

        self.assertGreater(last_updated2, last_updated1)

    def test_domain_update_wrong_token(self):
        service_data = {'name': 'ownCloud', 'type': '_http._tcp', 'port': 10000, 'url': None}
        update_data = {'token': create_token(), 'ip': '127.0.0.1', 'services': [service_data]}

        response = self.app.post('/domain/update', data=json.dumps(update_data))
        self.assertEqual(400, response.status_code)

    def test_domain_update_missing_port(self):
        email, password = self.create_active_user()

        user_domain = create_token()
        update_token = self.acquire_domain(email, password, user_domain)

        service_data = {'name': 'ownCloud', 'type': '_http._tcp', 'url': None}
        update_data = {'token': update_token, 'ip': '127.0.0.1', 'services': [service_data]}

        response = self.app.post('/domain/update', data=json.dumps(update_data))
        self.assertEqual(400, response.status_code)

    def test_domain_update_two_new_services(self):
        email, password = self.create_active_user()

        user_domain = create_token()
        update_token = self.acquire_domain(email, password, user_domain)

        services_data = [{'name': 'ownCloud', 'protocol': 'http', 'type': '_http._tcp', 'port': 10001, 'local_port': None, 'url': None},
                         {'name': 'SSH', 'protocol': 'https', 'type': '_http._tcp', 'port': 10002, 'local_port': None, 'url': None}]
        update_data = {'token': update_token, 'ip': '127.0.0.1', 'services': services_data}

        response = self.app.post('/domain/update', data=json.dumps(update_data))
        self.assertEqual(200, response.status_code)

        self.check_domain(update_token, {'ip': '127.0.0.1', 'user_domain': user_domain, 'services': services_data})

    def test_domain_update_services_change_ports(self):
        email, password = self.create_active_user()

        user_domain = create_token()
        update_token = self.acquire_domain(email, password, user_domain)


        services_data = [{'name': 'ownCloud', 'protocol': 'http', 'type': '_http._tcp', 'port': 10001, 'url': None},
                         {'name': 'SSH', 'protocol': 'https', 'type': '_tcp', 'port': 10002, 'url': None}]
        update_data = {'token': update_token, 'ip': '127.0.0.1', 'services': services_data}
        response = self.app.post('/domain/update', data=json.dumps(update_data))
        self.assertEqual(200, response.status_code)
        # self.check_domain(update_token, {'ip': '127.0.0.1', 'user_domain': user_domain, 'services': services_data})

        services_data = [{'name': 'ownCloud', 'protocol': 'http', 'type': '_http._tcp', 'port': 10002, 'url': None},
                         {'name': 'SSH', 'protocol': 'https', 'type': '_tcp', 'port': 10001, 'url': None}]
        update_data = {'token': update_token, 'ip': '127.0.0.1', 'services': services_data}
        response = self.app.post('/domain/update', data=json.dumps(update_data))
        self.assertEqual(200, response.status_code)

        # self.check_domain(update_token, {'ip': '127.0.0.1', 'user_domain': user_domain, 'services': services_data})

    def test_domain_update_service_removed(self):
        email, password = self.create_active_user()

        user_domain = create_token()
        update_token = self.acquire_domain(email, password, user_domain)

        services_data = [{'name': 'ownCloud', 'protocol': 'http', 'type': '_http._tcp', 'port': 10001, 'local_port': None, 'url': None},
                         {'name': 'SSH', 'protocol': 'https', 'type': '_http._tcp', 'port': 10002, 'local_port': None, 'url': None}]
        update_data = {'token': update_token, 'ip': '127.0.0.1', 'services': services_data}

        response = self.app.post('/domain/update', data=json.dumps(update_data))
        self.assertEqual(200, response.status_code)

        services_data = [{'name': 'ownCloud', 'protocol': 'http', 'type': '_http._tcp', 'port': 10001, 'local_port': None, 'url': None}]
        update_data = {'token': update_token, 'ip': '127.0.0.1', 'services': services_data}

        response = self.app.post('/domain/update', data=json.dumps(update_data))

        self.assertEqual(200, response.status_code)

        self.check_domain(update_token, {'ip': '127.0.0.1', 'user_domain': user_domain, 'services': services_data})

    def test_domain_update_service_updated(self):
        email, password = self.create_active_user()

        user_domain = create_token()
        update_token = self.acquire_domain(email, password, user_domain)

        services_data = [{'name': 'ownCloud', 'protocol': 'http', 'type': '_http._tcp', 'port': 10001, 'local_port': 80, 'url': None},
                         {'name': 'SSH', 'protocol': 'https', 'type': '_http._tcp', 'port': 10002, 'local_port': 81, 'url': None}]
        update_data = {'token': update_token, 'ip': '127.0.0.1', 'services': services_data}

        response = self.app.post('/domain/update', data=json.dumps(update_data))
        self.assertEqual(200, response.status_code)

        services_data = [{'name': 'ownCloud', 'protocol': 'http', 'type': '_http._tcp', 'port': 10001, 'local_port': 80, 'url': None},
                         {'name': 'SSH', 'protocol': 'https', 'type': '_http._tcp', 'port': 10003, 'local_port': 82, 'url': None}]
        update_data = {'token': update_token, 'ip': '127.0.0.1', 'services': services_data}

        response = self.app.post('/domain/update', data=json.dumps(update_data))

        self.assertEqual(200, response.status_code)

        self.check_domain(update_token, {'ip': '127.0.0.1', 'user_domain': user_domain, 'services': services_data})

    def test_domain_update_ip_changed(self):
        email, password = self.create_active_user()

        user_domain = create_token()
        update_token = self.acquire_domain(email, password, user_domain)

        services_data = [{'name': 'ownCloud', 'protocol': 'http', 'type': '_http._tcp', 'port': 10001, 'local_port': None, 'url': None}]
        update_data = {'token': update_token, 'ip': '127.0.0.1', 'services': services_data}

        response = self.app.post('/domain/update', data=json.dumps(update_data))
        self.assertEqual(200, response.status_code)

        services_data = [{'name': 'ownCloud', 'protocol': 'http', 'type': '_http._tcp', 'port': 10001, 'local_port': None, 'url': None}]
        update_data = {'token': update_token, 'ip': '127.0.0.2', 'services': services_data}

        response = self.app.post('/domain/update', data=json.dumps(update_data))

        self.assertEqual(200, response.status_code)

        self.check_domain(update_token, {'ip': '127.0.0.2', 'user_domain': user_domain, 'services': services_data})

    def test_local_ip_changed(self):
        email, password = self.create_active_user()

        user_domain = create_token()
        update_token = self.acquire_domain(email, password, user_domain)

        services_data = [{'name': 'ownCloud', 'protocol': 'http', 'type': '_http._tcp', 'port': 10001, 'local_port': None, 'url': None}]
        update_data = {'token': update_token, 'ip': '127.0.0.1', 'local_ip': '192.168.1.5', 'services': services_data}

        response = self.app.post('/domain/update', data=json.dumps(update_data))
        self.assertEqual(200, response.status_code)

        services_data = [{'name': 'ownCloud', 'protocol': 'http', 'type': '_http._tcp', 'port': 10001, 'local_port': None, 'url': None}]
        update_data = {'token': update_token, 'ip': '127.0.0.2', 'local_ip': '192.168.1.6', 'services': services_data}

        response = self.app.post('/domain/update', data=json.dumps(update_data))

        self.assertEqual(200, response.status_code)

        self.check_domain(update_token, {'ip': '127.0.0.2', 'local_ip': '192.168.1.6', 'user_domain': user_domain, 'services': services_data})

    def test_domain_update_server_side_client_ip(self):
        email, password = self.create_active_user()

        user_domain = create_token()
        update_token = self.acquire_domain(email, password, user_domain)

        services_data = [{'name': 'ownCloud', 'protocol': 'http', 'type': '_http._tcp', 'port': 10001, 'local_port': None, 'url': None}]
        update_data = {'token': update_token, 'services': services_data}

        response = self.app.post('/domain/update', data=json.dumps(update_data), environ_base={'REMOTE_ADDR': '127.0.0.1'})
        self.assertEqual(200, response.status_code)

        self.check_domain(update_token, {'ip': '127.0.0.1', 'user_domain': user_domain, 'services': services_data})

