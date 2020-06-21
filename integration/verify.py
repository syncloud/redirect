import json
from subprocess import check_output
from urlparse import urlparse

import pytest
from os.path import dirname, join
from syncloudlib.integration.hosts import add_host_alias_by_ip
import requests
import db
import uuid
import smtp

DIR = dirname(__file__)


@pytest.fixture(scope="session")
def module_setup(request, log_dir, artifact_dir):
    def module_teardown():
        check_output('cp /var/log/apache2/error.log {0}'.format(log_dir), shell=True)
        check_output('cp /var/log/apache2/redirect_rest-error.log {0}'.format(log_dir), shell=True)
        check_output('cp /var/log/apache2/redirect_rest-access.log {0}'.format(log_dir), shell=True)
        check_output('cp /var/log/apache2/redirect_ssl_rest-error.log {0}'.format(log_dir), shell=True)
        check_output('cp /var/log/apache2/redirect_ssl_rest-access.log {0}'.format(log_dir), shell=True)
        check_output('cp /var/log/apache2/redirect_ssl_web-access.log {0}'.format(log_dir), shell=True)
        check_output('cp /var/log/apache2/redirect_ssl_web-error.log {0}'.format(log_dir), shell=True)
        check_output('ls -la /var/log/apache2 > {0}/var.log.apache2.ls.log'.format(log_dir), shell=True)
        check_output('ls -la /var/log > {0}/var.log.ls.log'.format(log_dir), shell=True)

        check_output('chmod -R a+r {0}'.format(artifact_dir), shell=True)
        db.recreate()

    request.addfinalizer(module_teardown)


def create_token():
    return unicode(uuid.uuid4().hex)


def get_token(body):
    link_index = body.find('http://')
    link = body[link_index:].split(' ')[0].strip()
    parts = urlparse(link)
    token = parts.query.replace('token=', '')
    return token


def create_active_user(self):
    smtp.clear()
    email = '@mail.com'
    password = 'pass123456'
    self.www.post('/user/create', data={'email': email, 'password': password})
    activate_token = self.get_token(self.smtp.emails()[0])
    self.app.get('/user/activate', query_string={'token': activate_token})
    self.smtp.clear()
    return email, password


def test_start(module_setup, domain):
    add_host_alias_by_ip('app', 'www', '127.0.0.1', domain)
    add_host_alias_by_ip('app', 'api', '127.0.0.1', domain)


def test_index(domain):
    response = requests.get('https://www.{0}'.format(domain), allow_redirects=False, verify=False)
    assert response.status_code == 200, response.text


def test_user_create_special_symbols_in_password(self):
    email = 'symbols_in_password@mail.com'
    response = requests.post('/user/create', data={'email': email, 'password': r'pass12& ^%"'})
    assert response.status_code == 200
    assert smtp.emails() == 0


def test_user_create_success(domain, log_dir):
    email = 'test@syncloud.test'
    password = 'pass123456'
    response = requests.post('https://www.{0}/api/user/create'.format(domain), data={'email': email, 'password': password}, verify=False)
    assert response.status_code == 200, response.text
    assert smtp.emails() > 0
    activate_token = get_token(smtp.emails()[0])
    requests.get('/user/activate', query_string={'token': activate_token})
    smtp.clear()
    response = requests.get('/user/get', query_string={'email': email, 'password': password})
    assert response.status_code== 200, response.text
