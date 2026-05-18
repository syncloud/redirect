from os import environ
from os.path import join
from subprocess import check_output

import pytest
import requests

import db

TMP_DIR = '/tmp/syncloud'


@pytest.fixture(scope="session")
def module_setup(request, log_dir, artifact_dir, device):
    def module_teardown():
        device.run_ssh('cp /var/log/apache2/error.log {0}'.format(TMP_DIR), throw=False)
        device.run_ssh('cp /var/log/apache2/redirect_rest-error.log {0}'.format(TMP_DIR), throw=False)
        device.run_ssh('cp /var/log/apache2/redirect_rest-access.log {0}'.format(TMP_DIR), throw=False)
        device.run_ssh('cp /var/log/apache2/redirect_ssl_rest-error.log {0}'.format(TMP_DIR), throw=False)
        device.run_ssh('cp /var/log/apache2/redirect_ssl_rest-access.log {0}'.format(TMP_DIR), throw=False)
        device.run_ssh('cp /var/log/apache2/redirect_ssl_web-access.log {0}'.format(TMP_DIR), throw=False)
        device.run_ssh('cp /var/log/apache2/redirect_ssl_web-error.log {0}'.format(TMP_DIR), throw=False)
        device.run_ssh('ls -la /var/log/apache2 > {0}/var.log.apache2.ls.log'.format(TMP_DIR), throw=False)
        device.run_ssh('ls -la /var/log > {0}/var.log.ls.log'.format(TMP_DIR), throw=False)
        device.run_ssh('ls -la /var/run > {0}/var.run.ls.log'.format(TMP_DIR), throw=False)
        device.run_ssh('journalctl | tail -500 > {0}/journalctl.log'.format(TMP_DIR), throw=False)
        device.run_ssh('cp /var/log/syslog {0}/syslog.log'.format(TMP_DIR), throw=False)
        check_output("mysql --host=mysql --user=root --password=root redirect -e 'select * from user' > {0}/db-user.log || true".format(artifact_dir), shell=True)
        check_output("mysql --host=mysql --user=root --password=root redirect -e 'select * from action' > {0}/db-action.log || true".format(artifact_dir), shell=True)
        check_output("mysql --host=mysql --user=root --password=root redirect -e 'select * from domain' > {0}/db-domain.log || true".format(artifact_dir), shell=True)

        device.scp_from_device('{0}/*'.format(TMP_DIR), artifact_dir)
        check_output('chmod -R a+r {0}'.format(artifact_dir), shell=True)
        db.recreate()

    request.addfinalizer(module_teardown)


def test_start(module_setup, device, build_number):
    device.run_ssh('mkdir {0}'.format(TMP_DIR))
    device.run_ssh("apt-get update --allow-releaseinfo-change")
    device.run_ssh(
        "apt-get install -y default-mysql-client default-libmysqlclient-dev apache2 libapache2-mod-wsgi-py3 openssl > {0}/apt.log".format(
            TMP_DIR))
    device.scp_to_device("fakecertificate.sh", "/")
    device.run_ssh("/fakecertificate.sh")
    device.scp_to_device("../artifact/redirect-{0}.tar.gz".format(build_number), "/")
    device.scp_to_device("../ci/deploy", "/")
    # test clean deploy
    check_output("mysql --host=mysql --user=root --password=root -e 'drop DATABASE redirect'", shell=True)
    device.run_ssh("cd / && /deploy {0} integration syncloud.test > {1}/deploy.log 2>&1".format(build_number, TMP_DIR))
    device.run_ssh("sed -i 's#@access_key_id@#{0}#g' /var/www/redirect/secret.cfg".format(environ['access_key_id']),
                   debug=False)
    device.run_ssh(
        "sed -i 's#@secret_access_key@#{0}#g' /var/www/redirect/secret.cfg".format(environ['secret_access_key']),
        debug=False)
    device.run_ssh("sed -i 's#@hosted_zone_id@#{0}#g' /var/www/redirect/secret.cfg".format(environ['hosted_zone_id']),
                   debug=False)
    device.run_ssh("systemctl restart redirect.api")
    device.run_ssh("systemctl restart redirect.www")


def test_index(domain, artifact_dir):
    response = requests.get('https://www.{0}'.format(domain), allow_redirects=False, verify=False)
    assert response.status_code == 200, response.text
    with open(join(artifact_dir, 'index.html.log'), 'w') as f:
        f.write(str(response.text))
