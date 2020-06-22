import uuid
from os.path import dirname
from subprocess import check_output
from urlparse import urlparse

import pytest
import requests
from syncloudlib.integration.hosts import add_host_alias_by_ip

import db
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


def test_start(module_setup, domain):
    add_host_alias_by_ip('app', 'www', '127.0.0.1', domain)
    add_host_alias_by_ip('app', 'api', '127.0.0.1', domain)


def test_index(domain):
    response = requests.get('https://www.{0}'.format(domain), allow_redirects=False, verify=False)
    assert response.status_code == 200, response.text


def test_user_create_special_symbols_in_password(domain):
    email = 'symbols_in_password@mail.com'
    response = requests.post('https://www.{0}/api/user/create'.format(domain),
                             data={'email': email, 'password': r'pass12& ^%"'},
                             verify=False)
    assert response.status_code == 200
    assert len(smtp.emails()) == 1
    smtp.clear()


def test_user_create_success(domain):
    email = 'test@syncloud.test'
    password = 'pass123456'
    response = requests.post('https://www.{0}/api/user/create'.format(domain),
                             data={'email': email, 'password': password}, verify=False)
    assert response.status_code == 200, response.text
    assert len(smtp.emails()) == 1
    activate_token = get_token(smtp.emails()[0])
    response  = requests.get('https://api.{0}/user/activate?token={1}'.format(domain, activate_token),
                 verify=False)
    assert response.status_code == 200, response.text
    smtp.clear()
    response = requests.get('https://api.{0}/user/get?email={1}&password={2}'.format(domain, email, password),
                            verify=False)
    assert response.status_code == 200, response.text
