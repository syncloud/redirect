import json
from subprocess import check_output
import pytest
from os.path import dirname, join
from syncloudlib.integration.hosts import add_host_alias_by_ip
import requests
import db
import uuid

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


def test_start(module_setup, domain):
    add_host_alias_by_ip('app', 'www', '127.0.0.1', domain)
    add_host_alias_by_ip('app', 'api', '127.0.0.1', domain)


def test_index(domain):
    response = requests.get('https://www.{0}'.format(domain), allow_redirects=False, verify=False)
    assert response.status_code == 200, response.text


def test_user_create_success(domain, log_dir):
    email = 'test@syncloud.test'
    response = requests.post('https://www.{0}/api/user/create'.format(domain), data={'email': email, 'password': 'pass123456'}, verify=False)
    assert response.status_code == 200, response.text
    response = requests.get('http://mail:8025/api/v1/messages')
    with open(join(log_dir, 'mail.user.create.messages.log'), 'w') as f:
        f.write(str(response.text).replace(',', '\n'))
    assert response.status_code == 200, response.text
    assert len(json.loads(response.text)) == 1, response.text
    response = requests.delete('http://mail:8025/api/v1/messages')
    assert response.status_code == 200, response.text

