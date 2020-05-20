import os
from subprocess import check_output
import pytest
from syncloudlib.integration.hosts import add_host_alias_by_ip
import requests

DIR = dirname(__file__)


@pytest.fixture(scope="session")
def module_setup(request):
    def module_teardown():
        log_dir = DIR + '../artifact/log'
        os.mkdir(log_dir)
        check_output('cp /var/log/apache/error.log {0}'.format(log_dir))

    request.addfinalizer(module_teardown)


def test_start(module_setup):
    add_host_alias_by_ip('www', 'syncloud.it', '127.0.0.1')

def test_index():
    response = requests.get('https://www.syncloud.it', allow_redirects=False, verify=False)
    assert response.status_code == 200, response.text


