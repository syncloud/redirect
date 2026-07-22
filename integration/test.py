import json
import os
import ssl
import subprocess
import tarfile
import tempfile
import threading
import time
import urllib.request
from http.server import BaseHTTPRequestHandler, HTTPServer
from os.path import dirname, join

import pytest
import requests
from syncloudlib.integration.hosts import add_host_alias

import smtp
import api

DIR = dirname(__file__)


def get_domain(update_token, domain):
    response = requests.get('https://api.{0}/domain/get'.format(domain),
                            params={'token': update_token}, verify=False)
    assert response.status_code == 200
    assert response.text is not None
    response_data = json.loads(response.text)
    return response_data['data']


def test_unauthenticated(domain):
    response = requests.get('https://www.{0}/api/user'.format(domain), allow_redirects=False, verify=False)
    assert response.headers['Content-Type'] == 'application/json'
    assert response.status_code == 401, response.text


def test_user_create_special_symbols_in_password(domain):
    email = 'symbols_in_password@mail.com'
    response = requests.post('https://www.{0}/api/user/create'.format(domain),
                             json={'email': email, 'password': r'pass12& ^%"'},
                             verify=False)
    assert response.status_code == 200, response.text
    assert len(smtp.emails()) == 1
    smtp.clear()


def create_user(domain, email, password, artifact_dir):
    response = requests.post('https://www.{0}/api/user/create'.format(domain),
                             json={'email': email, 'password': password}, verify=False)
    assert response.status_code == 200, response.text

    activate_user(domain, artifact_dir)
    response = requests.get('https://api.{0}/user/get'.format(domain),
                            params={'email': email, 'password': password},
                            verify=False)
    assert response.status_code == 200, response.text

    response_data = json.loads(response.text)
    user_data = response_data['data']
    update_token = user_data["update_token"]

    return update_token


def test_create_user_api_for_mobile_app(domain, artifact_dir):
    email = 'mobile_create_user@syncloud.test'
    password = 'pass123456'
    response = requests.post('https://api.{0}/user/create'.format(domain),
                             data={'email': email, 'password': password}, verify=False)
    assert response.status_code == 200, response.text

    activate_user(domain, artifact_dir)

    response = requests.get('https://api.{0}/user/get'.format(domain),
                            params={'email': email, 'password': password},
                            verify=False)
    assert response.status_code == 200, response.text


def test_create_user_api_for_mobile_app_v2(domain, artifact_dir):
    email = 'mobile_create_user_v2@syncloud.test'
    password = 'pass123456'
    response = requests.post('https://api.{0}/user/create_v2'.format(domain),
                             json={'email': email, 'password': password}, verify=False)
    assert response.status_code == 200, response.text

    activate_user(domain, artifact_dir)

    response = requests.get('https://api.{0}/user/get'.format(domain),
                            params={'email': email, 'password': password},
                            verify=False)
    assert response.status_code == 200, response.text


def activate_user(domain, artifact_dir):
    assert len(smtp.emails(artifact_dir)) == 1
    activate_token = smtp.get_token(smtp.emails()[0])
    response = requests.post('https://www.{0}/api/user/activate'.format(domain),
                             json={'token': activate_token},
                             verify=False)
    assert response.status_code == 200, (response.text, activate_token)
    smtp.clear()


def test_user_create_success(domain, artifact_dir):
    create_user(domain, 'test@syncloud.test', 'pass123456', artifact_dir)


def test_user_create_existing_case_difference(domain, artifact_dir):
    email1 = 'case_test@syncloud.test'
    email2 = 'Case_test@syncloud.test'
    create_user(domain, email1, 'pass123456', artifact_dir)
    response = requests.post('https://www.{0}/api/user/create'.format(domain),
                             json={'email': email2, 'password': 'pass123456'}, verify=False)
    assert response.status_code == 400, response.text
    assert "already registered" in response.text, response.text


def test_get_user_data(domain, artifact_dir):
    email = 'test-get-user-data@syncloud.test'
    password = 'pass123456'
    user_token = create_user(domain, email, password, artifact_dir)

    user_domain = "test-get-user-data"
    update_token = api.domain_acquire(domain, '{}.{}'.format(user_domain, domain), email, password)

    update_data = {
        'token': update_token,
        'ip': '192.192.1.1',
        'web_protocol': 'http',
        'web_local_port': 80,
        'web_port': 10000
    }

    response = requests.post('https://api.{0}/domain/update'.format(domain),
                             json=update_data,
                             verify=False)
    assert response.status_code == 200, response.text

    response_data = json.loads(response.text)
    assert response_data['success'], response.text

    response = requests.get('https://api.{0}/user/get'.format(domain),
                            params={'email': email, 'password': password},
                            verify=False)

    response_data = json.loads(response.text)
    user_data = response_data['data']

    expected = {
        'active': True,
        'email': email,
        'unsubscribed': False,
        'update_token': user_token,
        'timestamp': user_data["timestamp"],
        'domains': [{
            'user_domain': user_domain,
            'web_local_port': 80,
            'web_port': 10000,
            'web_protocol': 'http',
            'ip': '192.192.1.1',
            'map_local_address': False,
            'device_mac_address': '00:00:00:00:00:00',
            'device_name': 'some-device',
            'device_title': 'Some Device',
            'last_update': user_data["domains"][0]["last_update"],
            'update_token': update_token,
            'name': 'test-get-user-data.syncloud.test'
        }]
    }

    assert expected == user_data


def test_get_user_no_domains(domain, artifact_dir):
    email = 'test_get_user_no_domains@syncloud.test'
    password = 'pass123456'
    user_token = create_user(domain, email, password, artifact_dir)

    response = requests.get('https://api.{0}/user/get'.format(domain),
                            params={'email': email, 'password': password},
                            verify=False)

    response_data = json.loads(response.text)
    user_data = response_data['data']

    expected = {
        'active': True,
        'email': email,
        'unsubscribed': False,
        'update_token': user_token,
        'timestamp': user_data["timestamp"],
        'domains': []
    }

    assert expected == user_data


def test_free_domain_availability(domain, artifact_dir):
    email = 'test-domain-availability@syncloud.test'
    password = 'pass123456'
    create_user(domain, email, password, artifact_dir)

    full_domain = 'domain-availability.syncloud.test'
    request = {
        'domain': full_domain,
        'email': email,
        'password': password,
    }

    response = requests.post('https://api.{0}/domain/availability'.format(domain),
                             json=request,
                             verify=False)
    assert response.status_code == 200, response.text

    user_domain = 'domain-availability'
    api.domain_acquire(domain, '{}.{}'.format(user_domain, domain), email, password)

    response = requests.post('https://api.{0}/domain/availability'.format(domain),
                             json=request,
                             verify=False)
    assert response.status_code == 200, response.text

    email = 'test_domain_availability_other@syncloud.test'
    password = 'pass123456'
    create_user(domain, email, password, artifact_dir)
    request = {
        'domain': full_domain,
        'email': email,
        'password': password,
    }
    response = requests.post('https://api.{0}/domain/availability'.format(domain),
                             json=request,
                             verify=False)
    assert response.status_code == 400, response.text


def test_user_reset_password_sent_mail(domain, artifact_dir):
    email = 'test_user_reset_password_sent_mail@syncloud.test'
    password = 'pass123456'
    create_user(domain, email, password, artifact_dir)

    response = requests.post('https://www.{0}/api/user/reset_password'.format(domain),
                             json={'email': email}, verify=False)
    assert response.status_code == 200, response.text

    assert len(smtp.emails()) > 0, 'Server should send email with link to reset password'
    token = smtp.get_token(smtp.emails()[0])
    smtp.clear()
    assert token is not None


def test_user_reset_password_set_new(domain, artifact_dir):
    email = 'test_user_reset_password_set_new@syncloud.test'
    password = 'pass123456'
    create_user(domain, email, password, artifact_dir)

    requests.post('https://www.{0}/api/user/reset_password'.format(domain), json={'email': email},
                  verify=False)
    email_body = smtp.emails(artifact_dir)[0]
    token = smtp.get_token(email_body)

    smtp.clear()

    new_password = 'new_password'
    response = requests.post('https://www.{0}/api/user/set_password'.format(domain),
                             json={'token': token, 'password': new_password},
                             verify=False)
    assert response.status_code == 200, (response.text, token, email_body)

    assert len(smtp.emails(artifact_dir)) > 0, 'Server should send email when setting new password'

    response = requests.get('https://api.{0}/user/get'.format(domain),
                            params={'email': email, 'password': new_password},
                            verify=False)
    assert response.status_code == 200, response.text
    smtp.clear()


def test_user_reset_password_set_with_old_token(domain, artifact_dir):
    email = 'test_user_reset_password_set_with_old_token@syncloud.test'
    password = 'pass123456'
    create_user(domain, email, password, artifact_dir)

    requests.post('https://www.{0}/api/user/reset_password'.format(domain), json={'email': email},
                  verify=False)
    token_old = smtp.get_token(smtp.emails()[0])

    smtp.clear()

    requests.post('https://www.{0}/api/user/reset_password'.format(domain), json={'email': email},
                  verify=False)
    token = smtp.get_token(smtp.emails()[0])
    smtp.clear()

    new_password = 'new_password'
    response = requests.post('https://www.{0}/api/user/set_password'.format(domain),
                             json={'token': token_old, 'password': new_password},
                             verify=False)
    assert response.status_code == 400, response.text
    smtp.clear()


def test_user_reset_password_set_twice(domain, artifact_dir):
    email = 'test_user_reset_password_set_twice@syncloud.test'
    password = 'pass123456'
    create_user(domain, email, password, artifact_dir)

    requests.post('https://www.{0}/api/user/reset_password'.format(domain), json={'email': email},
                  verify=False)
    token = smtp.get_token(smtp.emails()[0])
    smtp.clear()

    new_password = 'new_password'
    response = requests.post('https://www.{0}/api/user/set_password'.format(domain),
                             json={'token': token, 'password': new_password},
                             verify=False)
    assert response.status_code == 200, response.text

    new_password = 'new_password2'
    response = requests.post('https://www.{0}/api/user/set_password'.format(domain),
                             json={'token': token, 'password': new_password},
                             verify=False)
    assert response.status_code == 400, response.text
    smtp.clear()


def test_domain_new(domain, artifact_dir):
    email = 'test_domain_new@syncloud.test'
    password = 'pass123456'
    create_user(domain, email, password, artifact_dir)

    user_domain = "test-domain-new"
    acquire_data = dict(
        user_domain=user_domain,
        device_mac_address='00:00:00:00:00:00',
        device_name='my-super-board',
        device_title='My Super Board',
        email=email,
        password=password)
    response = requests.post('https://api.{0}/domain/acquire'.format(domain), data=acquire_data,
                             verify=False)

    assert response.status_code == 200
    domain_data = json.loads(response.text)

    update_token = domain_data['update_token']
    assert update_token is not None

    expected_data = {
        'update_token': update_token,
        'user_domain': user_domain,
        'device_mac_address': '00:00:00:00:00:00',
        'device_name': 'my-super-board',
        'map_local_address': False,
        'device_title': 'My Super Board',
        'name': 'test-domain-new.syncloud.test'
    }

    data = get_domain(update_token, domain)
    data.pop('last_update', None)
    assert expected_data == data


def test_domain_new_v2(domain, artifact_dir):
    email = 'test_domain_new_v2@syncloud.test'
    password = 'pass123456'
    create_user(domain, email, password, artifact_dir)

    request_domain = "test-domain-new-v2.syncloud.test"
    acquire_data = dict(
        domain=request_domain,
        device_mac_address='00:00:00:00:00:00',
        device_name='my-super-board',
        device_title='My Super Board',
        email=email,
        password=password)
    response = requests.post('https://api.{0}/domain/acquire_v2'.format(domain), json=acquire_data,
                             verify=False)

    assert response.status_code == 200
    domain_data = json.loads(response.text)['data']

    update_token = domain_data['update_token']
    assert update_token is not None

    expected_data = {
        'update_token': update_token,
        'user_domain': 'test-domain-new-v2',
        'device_mac_address': '00:00:00:00:00:00',
        'device_name': 'my-super-board',
        'device_title': 'My Super Board',
        'map_local_address': False,
        'name': request_domain
    }

    data = get_domain(update_token, domain)
    data.pop('last_update', None)
    assert expected_data == data

def test_domain_lower_case(domain, artifact_dir):
    email = 'test_domain_lower_case@syncloud.test'
    password = 'pass123456'
    create_user(domain, email, password, artifact_dir)

    request_domain = "test-domain-LOWER-case.syncloud.test"
    acquire_data = dict(
        domain=request_domain,
        device_mac_address='00:00:00:00:00:00',
        device_name='my-super-board',
        device_title='My Super Board',
        email=email,
        password=password)
    response = requests.post('https://api.{0}/domain/acquire_v2'.format(domain), json=acquire_data,
                             verify=False)

    assert response.status_code == 200
    domain_data = json.loads(response.text)['data']

    update_token = domain_data['update_token']
    assert update_token is not None

    expected_data = {
        'update_token': update_token,
        'user_domain': 'test-domain-lower-case',
        'device_mac_address': '00:00:00:00:00:00',
        'device_name': 'my-super-board',
        'device_title': 'My Super Board',
        'map_local_address': False,
        'name': 'test-domain-lower-case.syncloud.test'
    }

    data = get_domain(update_token, domain)
    data.pop('last_update', None)
    assert expected_data == data


def test_domain_existing(domain, artifact_dir):
    email_1 = 'test_domain_existing_@syncloud.test'
    password_1 = 'pass123456_'
    create_user(domain, email_1, password_1, artifact_dir)

    user_domain = "test-domain-existing"
    acquire_data = dict(
        user_domain=user_domain,
        device_mac_address='00:00:00:00:00:00',
        device_name='my-super-board',
        device_title='My Super Board',
        email=email_1,
        password=password_1)
    response = requests.post('https://api.{0}/domain/acquire'.format(domain), data=acquire_data,
                             verify=False)
    domain_data = json.loads(response.text)
    update_token = domain_data['update_token']

    email_2 = 'test_domain_existing@syncloud.test'
    password_2 = 'pass123456'
    create_user(domain, email_2, password_2, artifact_dir)
    acquire_data = dict(
        user_domain=user_domain,
        device_mac_address='00:00:00:00:00:11',
        device_name='other-board',
        device_title='Other Board',
        email=email_2,
        password=password_2)
    response = requests.post('https://api.{0}/domain/acquire'.format(domain), data=acquire_data,
                             verify=False)

    assert response.status_code == 400

    expected_data = {
        'update_token': update_token,
        'user_domain': user_domain,
        'device_mac_address': '00:00:00:00:00:00',
        'device_name': 'my-super-board',
        'device_title': 'My Super Board',
        'map_local_address': False,
        'name': 'test-domain-existing.syncloud.test'
    }

    data = get_domain(update_token, domain)
    data.pop('last_update', None)
    assert expected_data == data


def test_domain_twice(domain, artifact_dir):
    email = 'test_domain_twice@syncloud.test'
    password = 'pass123456_'
    create_user(domain, email, password, artifact_dir)

    user_domain = "test-domain-twice"
    acquire_data = dict(
        user_domain=user_domain,
        device_mac_address='00:00:00:00:00:00',
        device_name='my-super-board',
        device_title='My Super Board',
        email=email,
        password=password)
    response = requests.post('https://api.{0}/domain/acquire'.format(domain), data=acquire_data,
                             verify=False)
    domain_data = json.loads(response.text)
    update_token1 = domain_data['update_token']

    acquire_data = dict(
        user_domain=user_domain,
        device_mac_address='00:00:00:00:00:11',
        device_name='my-super-board-2',
        device_title='My Super Board 2',
        email=email,
        password=password)
    response = requests.post('https://api.{0}/domain/acquire'.format(domain), data=acquire_data,
                             verify=False)
    assert response.status_code == 200
    domain_data = json.loads(response.text)
    update_token2 = domain_data['update_token']

    assert update_token1 != update_token2

    expected_data = {
        'user_domain': user_domain,
        'update_token': update_token2,
        'device_mac_address': '00:00:00:00:00:11',
        'device_name': 'my-super-board-2',
        'device_title': 'My Super Board 2',
        'map_local_address': False,
        'name': 'test-domain-twice.syncloud.test'
    }

    data = get_domain(update_token2, domain)
    data.pop('last_update', None)
    assert expected_data == data


def test_domain_wrong_mac_address_format(domain, artifact_dir):
    email = 'test_domain_wrong_mac_address_format@syncloud.test'
    password = 'pass123456_'
    create_user(domain, email, password, artifact_dir)

    user_domain = "test-domain-wrong-mac-address-format"
    acquire_data = {
        'user_domain': user_domain,
        'device_mac_address': 'wrong_format',
        'device_name': 'my-super-board',
        'device_title': 'My Super Board',
        'email': email,
        'password': password
    }
    response = requests.post('https://api.{0}/domain/acquire'.format(domain), data=acquire_data,
                             verify=False)

    assert response.status_code == 400


def test_domain_update_date(domain, artifact_dir):
    email = 'test_domain_update_date@syncloud.test'
    password = 'pass123456'
    create_user(domain, email, password, artifact_dir)

    user_domain = "test-domain-update-date"
    update_token = api.domain_acquire(domain, '{}.{}'.format(user_domain, domain), email, password)
    api.domain_update(domain, update_token, '127.0.0.1')
    domain_info = get_domain(update_token, domain)
    last_updated1 = domain_info['last_update']
    time.sleep(1)
    api.domain_update(domain, update_token, '127.0.0.1')
    domain_info = get_domain(update_token, domain)
    last_updated2 = domain_info['last_update']

    assert last_updated2 > last_updated1


def test_domain_update_wrong_token(domain):
    update_token = 'test_domain_update_wrong_token'
    response = api.domain_update(domain, update_token, '127.0.0.1')
    assert response.status_code == 400, response.text


def test_domain_update_ip_changed(domain, artifact_dir):
    email = 'test_domain_update_ip_changed@syncloud.test'
    password = 'pass123456'
    create_user(domain, email, password, artifact_dir)
    user_domain = "test-domain-update-ip-changed"
    update_token = api.domain_acquire(domain, '{}.{}'.format(user_domain, domain), email, password)
    response = api.domain_update(domain, update_token, '127.0.0.1')
    assert response.status_code == 200
    response = api.domain_update(domain, update_token, '127.0.0.2')
    assert response.status_code == 200

    expected_data = {
        'update_token': update_token,
        'ip': '127.0.0.2',
        'user_domain': user_domain,
        'device_mac_address': '00:00:00:00:00:00',
        'device_name': 'some-device',
        'device_title': 'Some Device',
        'web_local_port': 80,
        'web_port': 10001,
        'web_protocol': 'https',
        'map_local_address': False,
        'name': 'test-domain-update-ip-changed.syncloud.test',
        'platform_version': '1'
    }

    domain_data = get_domain(update_token, domain)
    domain_data.pop('last_update', None)
    assert expected_data == domain_data


def test_domain_update_platform_version(domain, artifact_dir):
    email = 'test_domain_update_platform_version@syncloud.test'
    password = 'pass123456'
    create_user(domain, email, password, artifact_dir)
    user_domain = "test-domain-update-platform-version"

    update_token = api.domain_acquire(domain, '{}.{}'.format(user_domain, domain), email, password)
    response = api.domain_update(domain, update_token, '127.0.0.1', '366')
    assert response.status_code == 200

    expected_data = {
        'update_token': update_token,
        'platform_version': '366',
        'device_mac_address': '00:00:00:00:00:00',
        'device_name': 'some-device',
        'device_title': 'Some Device',
        'ip': '127.0.0.1',
        'user_domain': 'test-domain-update-platform-version',
        'web_local_port': 80,
        'web_port': 10001,
        'web_protocol': 'https',
        'map_local_address': False,
        'name': 'test-domain-update-platform-version.syncloud.test'
    }
    domain_data = get_domain(update_token, domain)
    domain_data.pop('last_update', None)
    assert expected_data == domain_data


def test_domain_update_local_ip_changed(domain, artifact_dir):
    email = 'test_domain_update_local_ip_changed@syncloud.test'
    password = 'pass123456'
    create_user(domain, email, password, artifact_dir)
    user_domain = "test-domain-update-local-ip-changed"

    update_token = api.domain_acquire(domain, '{}.{}'.format(user_domain, domain), email, password)

    update_data = {
        'token': update_token,
        'ip': '127.0.0.1',
        'local_ip': '192.168.1.5',
        'web_protocol': 'http',
        'web_port': 10001,
        'web_local_port': 80,
    }

    response = requests.post('https://api.{0}/domain/update'.format(domain), json=update_data,
                             verify=False)
    assert response.status_code == 200

    update_data = {
        'token': update_token,
        'ip': '127.0.0.2',
        'local_ip': '192.168.1.6',
        'web_protocol': 'http',
        'web_port': 10001,
        'web_local_port': 80,
    }

    response = requests.post('https://api.{0}/domain/update'.format(domain), json=update_data,
                             verify=False)

    assert response.status_code == 200

    expected_data = {
        'update_token': update_token,
        'ip': '127.0.0.2',
        'local_ip': '192.168.1.6',
        'user_domain': user_domain,
        'device_mac_address': '00:00:00:00:00:00',
        'device_name': 'some-device',
        'device_title': 'Some Device',
        'web_local_port': 80,
        'web_port': 10001,
        'web_protocol': 'http',
        'map_local_address': False,
        'name': 'test-domain-update-local-ip-changed.syncloud.test'
    }
    domain_data = get_domain(update_token, domain)
    domain_data.pop('last_update', None)
    assert expected_data == domain_data


def test_domain_update_server_side_client_ip(domain, artifact_dir):
    email = 'test_domain_update_server_side_client_ip@syncloud.test'
    password = 'pass123456'
    create_user(domain, email, password, artifact_dir)
    user_domain = "test-domain-update-server-side-client-ip"

    update_token = api.domain_acquire(domain, '{}.{}'.format(user_domain, domain), email, password)

    update_data = {
        'token': update_token,
        'web_protocol': 'http',
        'web_port': 10001,
        'web_local_port': 80,
    }

    response = requests.post('https://api.{0}/domain/update'.format(domain),
                             json=update_data,
                             verify=False)
    assert response.status_code == 200, response.text

    expected_data = {
        'update_token': update_token,
        'user_domain': user_domain,
        'web_protocol': 'http',
        'web_port': 10001,
        'web_local_port': 80,
        'device_mac_address': '00:00:00:00:00:00',
        'device_name': 'some-device',
        'device_title': 'Some Device',
        'map_local_address': False,
        'name': 'test-domain-update-server-side-client-ip.syncloud.test'
    }

    domain_data = get_domain(update_token, domain)
    domain_data.pop('last_update', None)
    domain_data.pop('ip', None)
    assert expected_data == domain_data


def test_domain_update_map_local_address(domain, artifact_dir):
    email = 'test_domain_update_map_local_address@syncloud.test'
    password = 'pass123456'
    create_user(domain, email, password, artifact_dir)

    user_domain = "test-domain-update-map-local-address"
    update_token = api.domain_acquire(domain, '{}.{}'.format(user_domain, domain), email, password)

    update_data = {
        'token': update_token,
        'ip': '108.108.108.108',
        'local_ip': '192.168.1.2',
        'map_local_address': True,
        'web_protocol': 'http',
        'web_port': 10001,
        'web_local_port': 80
    }

    response = requests.post('https://api.{0}/domain/update'.format(domain), json=update_data,
                             verify=False)
    assert response.status_code == 200

    expected_data = {
        'update_token': update_token,
        'ip': '108.108.108.108',
        'local_ip': '192.168.1.2',
        'map_local_address': True,
        'user_domain': user_domain,
        'device_mac_address': '00:00:00:00:00:00',
        'device_name': 'some-device',
        'device_title': 'Some Device',
        'web_protocol': 'http',
        'web_port': 10001,
        'web_local_port': 80,
        'name': 'test-domain-update-map-local-address.syncloud.test'
    }

    domain_data = get_domain(update_token, domain)
    domain_data.pop('last_update', None)
    assert expected_data == domain_data


def test_status(domain):
    response = requests.get('https://api.{0}/status'.format(domain), verify=False)
    assert response.status_code == 200
    assert 'OK' in response.text


def test_backup(device):
   device.run_ssh("/var/www/redirect/current/bin/redirectdb backup redirect redirect.sql")


def test_certbot(device, domain):
    device.run_ssh("mkdir /var/www/redirect/current/www/.well-known")
    device.run_ssh("echo OK > /var/www/redirect/current/www/.well-known/test")
    response = requests.get('http://api.{0}/.well-known/test'.format(domain), verify=False)
    assert response.status_code == 200
    assert 'OK' in response.text


def test_user_log(domain, artifact_dir):
    email = 'test_user_log@syncloud.test'
    password = 'pass123456'
    token = create_user(domain, email, password, artifact_dir)

    response = requests.post('https://api.{0}/user/log'.format(domain),
                             data={'token': token,
                                   'data': 'test_user_log',
                                   'include_support': False},
                             verify=False)
    assert response.status_code == 200, response.text

    assert len(smtp.emails()) > 0, 'Server should send email with log'
    email = smtp.emails()[0]
    smtp.clear()
    assert 'test_user_log' in email


def test_user_log_include_support(domain, artifact_dir):
    email = 'test_user_log_include_support@syncloud.test'
    password = 'pass123456'
    token = create_user(domain, email, password, artifact_dir)

    response = requests.post('https://api.{0}/user/log'.format(domain),
                             data={'token': token,
                                   'data': 'test_user_log_include_support',
                                   'include_support': True},
                             verify=False)
    assert response.status_code == 200, response.text

    assert len(smtp.emails()) > 0, 'Server should send email with log'
    email = smtp.emails()[0]
    smtp.clear()
    assert 'test_user_log_include_support' in email


def test_certbot_support(domain, artifact_dir):
    email = 'test_certbot_support@syncloud.test'
    password = 'pass123456'
    token = create_user(domain, email, password, artifact_dir)

    user_domain = 'test-certbot-support.{}'.format(domain)
    update_token = api.domain_acquire(domain, user_domain, email, password)

    response = requests.post('https://api.{0}/certbot/present'.format(domain),
                             json={
                                 'token': update_token,
                                 'fqdn': '_certbot.{}'.format(user_domain),
                                 'values': ['value1']
                             },
                             verify=False)
    assert response.status_code == 200, response.text

    response = requests.post('https://api.{0}/certbot/present'.format(domain),
                             json={
                                 'token': update_token,
                                 'fqdn': '_certbot.{}'.format(user_domain),
                                 'values': ['value1', 'value2']
                             },
                             verify=False)
    assert response.status_code == 200, response.text

    response = requests.post('https://api.{0}/certbot/cleanup'.format(domain),
                             json={
                                 'token': update_token,
                                 'fqdn': '_certbot.{}'.format(user_domain)
                             },
                             verify=False)
    assert response.status_code == 200, response.text

    response = requests.post('https://api.{0}/certbot/cleanup'.format(domain),
                             json={
                                 'token': update_token,
                                 'fqdn': '_certbot.{}'.format(user_domain)
                             },
                             verify=False)
    assert response.status_code == 200, response.text


RELAY_ADDRESS = '10.0.0.99'
FRP_VERSION = '0.70.0'
BACKEND_BODY = 'relay-backend-ok'
BACKEND_PORT = 18443
BIG_BODY = ('x' * 65536).encode()


@pytest.fixture(scope='session')
def frpc():
    url = 'https://github.com/fatedier/frp/releases/download/v{0}/frp_{0}_linux_amd64.tar.gz'.format(FRP_VERSION)
    work = tempfile.mkdtemp()
    tgz = join(work, 'frp.tgz')
    error = None
    for _ in range(5):
        try:
            urllib.request.urlretrieve(url, tgz)
            error = None
            break
        except Exception as e:
            error = e
            time.sleep(3)
    if error is not None:
        raise error
    with tarfile.open(tgz) as tar:
        tar.extractall(work)
    frpc_path = join(work, 'frp_{0}_linux_amd64'.format(FRP_VERSION), 'frpc')
    os.chmod(frpc_path, 0o755)
    return frpc_path


class RelayBackendHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        body = BIG_BODY if self.path == '/big' else BACKEND_BODY.encode()
        self.send_response(200)
        self.send_header('Content-Type', 'text/plain')
        self.send_header('Content-Length', str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def log_message(self, *args):
        pass


def relay_gen_cert(work_dir):
    cert = join(work_dir, 'backend.crt')
    key = join(work_dir, 'backend.key')
    subprocess.check_call([
        'openssl', 'req', '-x509', '-newkey', 'rsa:2048',
        '-keyout', key, '-out', cert, '-nodes', '-days', '1',
        '-subj', '/CN=relay-backend'])
    return cert, key


def relay_start_backend(work_dir):
    cert, key = relay_gen_cert(work_dir)
    httpd = HTTPServer(('127.0.0.1', BACKEND_PORT), RelayBackendHandler)
    context = ssl.SSLContext(ssl.PROTOCOL_TLS_SERVER)
    context.load_cert_chain(cert, key)
    httpd.socket = context.wrap_socket(httpd.socket, server_side=True)
    threading.Thread(target=httpd.serve_forever, daemon=True).start()
    return httpd


def relay_write_frpc_config(path, server_addr, server_name, token, domain_name):
    config = (
        'serverAddr = "{addr}"\n'
        'serverPort = 443\n'
        'transport.tls.enable = true\n'
        'transport.tls.serverName = "{sni}"\n'
        'transport.tls.disableCustomTLSFirstByte = true\n'
        'metadatas.token = "{token}"\n'
        '\n'
        '[[proxies]]\n'
        'name = "{domain}"\n'
        'type = "https"\n'
        'customDomains = ["{domain}"]\n'
        'localIP = "127.0.0.1"\n'
        'localPort = {port}\n'
    ).format(addr=server_addr, sni=server_name, token=token, domain=domain_name, port=BACKEND_PORT)
    with open(path, 'w') as f:
        f.write(config)


def relay_start_frpc(frpc_path, work_dir, server_addr, server_name, token, domain_name, tag):
    config_path = join(work_dir, 'frpc-{0}.toml'.format(tag))
    log_path = join(work_dir, 'frpc-{0}.log'.format(tag))
    relay_write_frpc_config(config_path, server_addr, server_name, token, domain_name)
    log = open(log_path, 'w')
    process = subprocess.Popen([frpc_path, '-c', config_path], stdout=log, stderr=subprocess.STDOUT)
    return process, log_path


def relay_fetch(domain_name):
    return requests.get('https://{0}/'.format(domain_name), verify=False, timeout=5)


def test_relay_valid_token_tunnels_traffic(domain, device_host, artifact_dir, frpc):
    user_domain = 'relaye2e'
    domain_name = '{0}.{1}'.format(user_domain, domain)
    email = 'relay_e2e@syncloud.test'
    password = 'pass123456'
    create_user(domain, email, password, artifact_dir)
    token = api.domain_acquire(domain, domain_name, email, password)
    add_host_alias(user_domain, device_host, domain)

    work_dir = tempfile.mkdtemp()
    backend = relay_start_backend(work_dir)
    process, log_path = relay_start_frpc(frpc, work_dir, device_host, 'relay.{0}'.format(domain), token, domain_name, 'valid')
    try:
        body = None
        for _ in range(30):
            try:
                response = relay_fetch(domain_name)
                if response.status_code == 200:
                    body = response.text
                    break
            except Exception:
                pass
            time.sleep(2)
        assert body == BACKEND_BODY, open(log_path).read()
    finally:
        process.terminate()
        backend.shutdown()


def test_relay_bad_token_rejected(domain, device_host, frpc):
    user_domain = 'relayneg'
    domain_name = '{0}.{1}'.format(user_domain, domain)
    add_host_alias(user_domain, device_host, domain)

    work_dir = tempfile.mkdtemp()
    bad_token = '00000000-0000-0000-0000-000000000000'
    process, log_path = relay_start_frpc(frpc, work_dir, device_host, 'relay.{0}'.format(domain), bad_token, domain_name, 'bad')
    try:
        time.sleep(10)
        got_backend = False
        try:
            got_backend = relay_fetch(domain_name).text == BACKEND_BODY
        except Exception:
            got_backend = False
        assert not got_backend, 'relay served traffic for a domain the token does not own'
    finally:
        process.terminate()


def test_relay_update_points_dns_at_relay(domain, artifact_dir):
    email = 'relay_dns@syncloud.test'
    password = 'pass123456'
    create_user(domain, email, password, artifact_dir)
    domain_name = 'relaydns.{0}'.format(domain)
    update_token = api.domain_acquire(domain, domain_name, email, password)

    response = requests.post('https://api.{0}/domain/update'.format(domain), json={
        'token': update_token,
        'ipv4_enabled': True,
        'relay': True,
        'web_protocol': 'https',
        'web_local_port': 443,
    }, verify=False)
    assert response.status_code == 200, response.text

    data = get_domain(update_token, domain)
    assert data['ip'] == RELAY_ADDRESS, data


def test_relay_monthly_limit_blocks_traffic(domain, device_host, artifact_dir, frpc):
    user_domain = 'relayquota'
    domain_name = '{0}.{1}'.format(user_domain, domain)
    email = 'relay_quota@syncloud.test'
    password = 'pass123456'
    create_user(domain, email, password, artifact_dir)
    token = api.domain_acquire(domain, domain_name, email, password)
    add_host_alias(user_domain, device_host, domain)

    work_dir = tempfile.mkdtemp()
    backend = relay_start_backend(work_dir)
    process, log_path = relay_start_frpc(frpc, work_dir, device_host, 'relay.{0}'.format(domain), token, domain_name, 'quota')
    try:
        up = False
        for _ in range(30):
            try:
                if relay_fetch(domain_name).status_code == 200:
                    up = True
                    break
            except Exception:
                pass
            time.sleep(2)
        assert up, open(log_path).read()

        time.sleep(4)

        for _ in range(5):
            try:
                requests.get('https://{0}/big'.format(domain_name), verify=False, timeout=5)
            except Exception:
                pass

        blocked = False
        for _ in range(20):
            try:
                if requests.get('https://{0}/big'.format(domain_name), verify=False, timeout=5).status_code != 200:
                    blocked = True
                    break
            except Exception:
                blocked = True
                break
            time.sleep(1)
        assert blocked, 'relay kept serving after exceeding the monthly limit\n' + open(log_path).read()
    finally:
        process.terminate()
        backend.shutdown()
