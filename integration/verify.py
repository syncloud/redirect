from subprocess import check_output
import pytest
from syncloudlib.integration.hosts import add_host_alias_by_ip
import requests


@pytest.fixture(scope="session")
def module_setup(request, log_dir, artifact_dir):
    def module_teardown():
        check_output('cp /var/log/apache2/error.log {0}'.format(log_dir), shell=True)
        check_output('chmod -R a+r {0}'.format(artifact_dir), shell=True)

    request.addfinalizer(module_teardown)


def test_start(module_setup):
    add_host_alias_by_ip('www', 'syncloud.it', '127.0.0.1')


def test_index():
    response = requests.get('https://www.syncloud.it', allow_redirects=False, verify=False)
    assert response.status_code == 200, response.text


