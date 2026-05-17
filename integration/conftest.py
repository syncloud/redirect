from os.path import dirname

from syncloudlib.integration.conftest import *
from syncloudlib.integration.hosts import add_host_alias

DIR = dirname(__file__)


@pytest.fixture(scope="session")
def project_dir():
    return join(dirname(__file__), '..')


@pytest.fixture(scope="session", autouse=True)
def host_aliases(device_host, domain):
    add_host_alias('api', device_host, domain)
