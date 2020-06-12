from subprocess import check_output
import pytest
from os.path import dirname
from syncloudlib.integration.hosts import add_host_alias_by_ip
import requests

DIR = dirname(__file__)


@pytest.fixture(scope="session")
def module_setup(request, log_dir, artifact_dir):
    def module_teardown():
        check_output('cp /var/log/apache2/error.log {0}'.format(log_dir), shell=True)
        check_output('cp /var/log/apache2/redirect_rest-error.log {0}/redirect_rest-error.log'.format(log_dir), shell=True)
        check_output('chmod -R a+r {0}'.format(artifact_dir), shell=True)

    request.addfinalizer(module_teardown)


def test_start(module_setup, domain):
    add_host_alias_by_ip('app', 'www', '127.0.0.1', domain)
    check_output('cp {0}/test_secret.cfg /var/www/redirect/current/redirect/secret.cfg'.format(DIR), shell=True)


def test_index(domain):
    response = requests.get('https://www.{0}'.format(domain), allow_redirects=False, verify=False)
    assert response.status_code == 200, response.text


