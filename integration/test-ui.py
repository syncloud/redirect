import time
from os.path import dirname
from subprocess import check_output

import pytest
from selenium.webdriver.common.keys import Keys
from syncloudlib.integration.hosts import add_host_alias_by_ip
from syncloudlib.integration.screenshots import screenshots
import smtp
import requests

import db

DIR = dirname(__file__)
DEVICE_USER="user@example.com"
DEVICE_PASSWORD="password"
TMP_DIR = '/tmp/syncloud'


@pytest.fixture(scope="session")
def module_setup(request, ui_mode, log_dir, artifact_dir, device):
    def module_teardown():
        device.run_ssh('ls -la /data/platform/backup > {0}/data.platform.backup.ls.log'.format(TMP_DIR), throw=False)

        device.run_ssh('cp /var/log/apache2/redirect_rest-error.log {0}/{1}-redirect_rest-error.log'.format(TMP_DIR, ui_mode), throw=False)
        device.run_ssh('cp /var/log/apache2/redirect_rest-access.log {0}/{1}-redirect_rest-access.log'.format(TMP_DIR, ui_mode), throw=False)
        device.run_ssh('cp /var/log/apache2/redirect_ssl_web-access.log {0}/{1}-redirect_ssl_web-access.log'.format(TMP_DIR, ui_mode), throw=False)
        device.run_ssh('cp /var/log/apache2/redirect_ssl_web-error.log {0}/{1}-redirect_ssl_web-error.log'.format(TMP_DIR, ui_mode), throw=False)
        device.scp_from_device('{0}/*'.format(TMP_DIR), artifact_dir)
        check_output('chmod -R a+r {0}'.format(artifact_dir), shell=True)

    request.addfinalizer(module_teardown)


def test_start(module_setup, device_host, domain):
    check_output('apt-get update', shell=True)
    check_output('apt-get install -y mysql-client', shell=True)
    add_host_alias_by_ip('app', 'www', device_host, domain)
    add_host_alias_by_ip('app', 'api', device_host, domain)
    db.recreate()


def test_login(driver, screenshot_dir, ui_mode, domain):
    driver.get("https://www.{0}".format(domain))
    screenshots(driver, screenshot_dir, 'index-' + ui_mode)
    time.sleep(10)


def test_register(driver, ui_mode, screenshot_dir, domain):
    driver.get("https://www.{0}/register.html".format(domain))
    screenshots(driver, screenshot_dir, 'register-' + ui_mode)
    user = driver.find_element_by_id("email")
    user.send_keys(DEVICE_USER)
    password = driver.find_element_by_id("password")
    password.send_keys(DEVICE_PASSWORD)
    screenshots(driver, screenshot_dir, 'register-credentials-' + ui_mode)
    password.send_keys(Keys.RETURN)
    time.sleep(2)
    screenshots(driver, screenshot_dir, 'register-progress-' + ui_mode)
    activate_token = smtp.get_token(smtp.emails()[0])
    response  = requests.get('https://api.{0}/user/activate?token={1}'.format(domain, activate_token),
                             verify=False)
    assert response.status_code == 200, response.text
    smtp.clear()


def test_main(driver, ui_mode, screenshot_dir, domain):
    driver.get("https://www.{0}".format(domain))
    screenshots(driver, screenshot_dir, 'login-' + ui_mode)
    user = driver.find_element_by_id("email")
    user.send_keys(DEVICE_USER)
    password = driver.find_element_by_id("password")
    password.send_keys(DEVICE_PASSWORD)
    screenshots(driver, screenshot_dir, 'login-credentials-' + ui_mode)
    password.send_keys(Keys.RETURN)
    time.sleep(2)
    screenshots(driver, screenshot_dir, 'login-progress-' + ui_mode)
    time.sleep(2)
    screenshots(driver, screenshot_dir, 'main-' + ui_mode)
    assert "You do not have any activated devices" in driver.page_source.encode("utf-8")

