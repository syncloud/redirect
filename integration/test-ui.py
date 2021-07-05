import json
import socket

import time
from os.path import dirname, join
from subprocess import check_output

import pytest
import requests
from selenium.webdriver.common.by import By
from selenium.webdriver.common.keys import Keys
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
from syncloudlib.integration.hosts import add_host_alias_by_ip
from syncloudlib.integration.screenshots import screenshots
import smtp

import db
import premium_account
import api

DIR = dirname(__file__)
DEVICE_USER = "user@example.com"
DEVICE_PASSWORD = "password"
TMP_DIR = '/tmp/syncloud'


@pytest.fixture(scope="session")
def module_setup(request, ui_mode, log_dir, artifact_dir, device):
    def module_teardown():
        device.run_ssh('mkdir {0}/{1}'.format(TMP_DIR, ui_mode), throw=False)
        device.run_ssh('ls -la /data/platform/backup > {0}/{1}/data.platform.backup.ls.log'.format(TMP_DIR, ui_mode), throw=False)
        device.run_ssh('cp /var/log/apache2/redirect_rest-error.log {0}/{1}/rest-error.log'.format(TMP_DIR, ui_mode), throw=False)
        device.run_ssh('cp /var/log/apache2/redirect_rest-access.log {0}/{1}/rest-access.log'.format(TMP_DIR, ui_mode), throw=False)
        device.run_ssh('cp /var/log/apache2/redirect_ssl_web-access.log {0}/{1}/web-access.log'.format(TMP_DIR, ui_mode), throw=False)
        device.run_ssh('cp /var/log/apache2/redirect_ssl_web-error.log {0}/{1}/web-error.log'.format(TMP_DIR, ui_mode), throw=False)
        device.run_ssh('journalctl | tail -500 > {0}/{1}/journalctl.log'.format(TMP_DIR, ui_mode), throw=False)
        device.scp_from_device('{0}/*'.format(TMP_DIR), artifact_dir)
        check_output("mysql --host=mysql --user=root --password=root redirect -e 'select * from action' > {0}/{1}/db-action.log || true".format(artifact_dir, ui_mode), shell=True)
        check_output("mysql --host=mysql --user=root --password=root redirect -e 'select * from user' > {0}/{1}/db-user.log".format(artifact_dir, ui_mode), shell=True)
        check_output("mysql --host=mysql --user=root --password=root redirect -e 'select * from domain' > {0}/{1}/db-domain.log".format(artifact_dir, ui_mode), shell=True)
        check_output('cp -R {0} {1}'.format(log_dir, artifact_dir), shell=True)
        check_output('chmod -R a+r {0}'.format(artifact_dir), shell=True)

    request.addfinalizer(module_teardown)


def test_start(module_setup, device_host, domain):
    add_host_alias_by_ip('app', 'api', device_host, domain)
    db.recreate()


def test_error(driver, screenshot_dir, ui_mode, domain):
    ip = socket.gethostbyname('www.syncloud.test')
    print('domain ip: ' + ip)
    driver.get("https://www.{0}/error".format(domain))
    wait_or_screenshot(driver, ui_mode, screenshot_dir, EC.presence_of_element_located((By.ID, 'error')))
    screenshots(driver, screenshot_dir, 'error-' + ui_mode)


def test_index(driver, screenshot_dir, ui_mode, domain):
    driver.get("https://www.{0}".format(domain))
    wait_or_screenshot(driver, ui_mode, screenshot_dir, EC.presence_of_element_located((By.ID, 'email')))
    screenshots(driver, screenshot_dir, 'index-' + ui_mode)


def test_register(driver, ui_mode, screenshot_dir):
    menu(driver, ui_mode, screenshot_dir, 'register')

    wait_or_screenshot(driver, ui_mode, screenshot_dir, EC.presence_of_element_located((By.ID, 'register_email')))
    screenshots(driver, screenshot_dir, 'register-' + ui_mode)
    email = driver.find_element_by_id('register_email')
    email.send_keys(DEVICE_USER)
    password = driver.find_element_by_id("register_password")
    password.send_keys(DEVICE_PASSWORD)
    screenshots(driver, screenshot_dir, 'register-credentials-' + ui_mode)
    password.send_keys(Keys.RETURN)
    wait_or_screenshot(driver, ui_mode, screenshot_dir, EC.presence_of_element_located((By.ID, 'email')))

    screenshots(driver, screenshot_dir, 'login-' + ui_mode)
    activate_url = smtp.get_activate_url(smtp.emails()[0])
    smtp.clear()
    driver.get(activate_url)
    print('activate_url: ' + activate_url)
    activated_status = "//div[text()='User was activated']"
    wait_or_screenshot(driver, ui_mode, screenshot_dir, EC.presence_of_element_located((By.XPATH, activated_status)))


def test_login(driver, ui_mode, screenshot_dir):
    menu(driver, ui_mode, screenshot_dir, 'login')

    wait_or_screenshot(driver, ui_mode, screenshot_dir, EC.presence_of_element_located((By.ID, 'email')))

    screenshots(driver, screenshot_dir, 'login-' + ui_mode)
    user = driver.find_element_by_id("email")
    user.send_keys(DEVICE_USER)
    password = driver.find_element_by_id("password")
    password.send_keys(DEVICE_PASSWORD)
    screenshots(driver, screenshot_dir, 'login-credentials-' + ui_mode)
    password.send_keys(Keys.RETURN)
    screenshots(driver, screenshot_dir, 'login-progress-' + ui_mode)
    wait_or_screenshot(driver, ui_mode, screenshot_dir, EC.visibility_of_element_located((By.ID, 'no_domains')))
    screenshots(driver, screenshot_dir, 'default-' + ui_mode)
    assert "You do not have any activated devices" in driver.page_source


def test_devices(domain, driver, ui_mode, screenshot_dir, artifact_dir):

    response = api.domain_acquire(domain, '{}.{}'.format(ui_mode, domain), DEVICE_USER, DEVICE_PASSWORD)
    acquire_response = json.loads(response.text)
    assert acquire_response['success'], response.text
    assert acquire_response['update_token'], response.text

    driver.get("https://www.{0}/api/domains".format(domain))
    with open(join(artifact_dir, '{}-api-domains.log'.format(ui_mode)), 'w') as f:
        f.write(str(driver.page_source.encode("utf-8")))
    driver.get("https://www.{0}".format(domain))

    menu(driver, ui_mode, screenshot_dir, 'devices')

    device_label = "//h3[text()='Some Device']"
    wait_or_screenshot(driver, ui_mode, screenshot_dir, EC.presence_of_element_located((By.XPATH, device_label)))
    by_xpath = driver.find_element_by_xpath(device_label)
    screenshots(driver, screenshot_dir, 'devices-' + ui_mode)
    assert by_xpath is not None


def test_password_reset(driver, ui_mode, screenshot_dir):
    menu(driver, ui_mode, screenshot_dir, 'logout')

    wait_or_screenshot(driver, ui_mode, screenshot_dir, EC.presence_of_element_located((By.ID, 'forgot')))
    forgot = driver.find_element_by_id('forgot')
    forgot.click()

    wait_or_screenshot(driver, ui_mode, screenshot_dir, EC.presence_of_element_located((By.ID, 'send')))

    email = driver.find_element_by_id('email')
    email.send_keys(DEVICE_USER)
    send = driver.find_element_by_id('send')
    send.click()

    reset_url = smtp.get_reset_url(smtp.emails()[0])
    smtp.clear()

    driver.get(reset_url)
    wait_or_screenshot(driver, ui_mode, screenshot_dir, EC.presence_of_element_located((By.ID, 'password')))
    password = driver.find_element_by_id('password')
    global DEVICE_PASSWORD
    DEVICE_PASSWORD = 'password1'
    password.send_keys(DEVICE_PASSWORD)
    reset = driver.find_element_by_id('reset')
    reset.click()

    menu(driver, ui_mode, screenshot_dir, 'login')

    wait_or_screenshot(driver, ui_mode, screenshot_dir, EC.presence_of_element_located((By.ID, 'email')))

    screenshots(driver, screenshot_dir, 'reset-login-' + ui_mode)
    user = driver.find_element_by_id("email")
    user.send_keys(DEVICE_USER)
    password = driver.find_element_by_id("password")
    password.send_keys(DEVICE_PASSWORD)
    screenshots(driver, screenshot_dir, 'reset-login-credentials-' + ui_mode)
    password.send_keys(Keys.RETURN)
    screenshots(driver, screenshot_dir, 'reset-login-progress-' + ui_mode)
    device_label = "//h3[text()='Some Device']"
    wait_or_screenshot(driver, ui_mode, screenshot_dir, EC.presence_of_element_located((By.XPATH, device_label)))
    by_xpath = driver.find_element_by_xpath(device_label)
    screenshots(driver, screenshot_dir, 'reset-devices-' + ui_mode)
    assert by_xpath is not None


def test_domain_delete(driver, ui_mode, screenshot_dir):
    domain_delete(driver, ui_mode, screenshot_dir, 'devices-removed')


def domain_delete(driver, ui_mode, screenshot_dir, screenshot):
    deactivate_xpath = "//div[contains(@class, 'panel')]//button[contains(text(), 'Deactivate')]"
    wait_or_screenshot(driver, ui_mode, screenshot_dir, EC.presence_of_element_located((By.XPATH, deactivate_xpath)))
    driver.find_element_by_xpath(deactivate_xpath).click()

    confirm_xpath = "//div[@id='delete_confirmation']//button[contains(text(), 'Yes')]"
    wait_or_screenshot(driver, ui_mode, screenshot_dir, EC.presence_of_element_located((By.XPATH, confirm_xpath)))
    driver.find_element_by_xpath(confirm_xpath).click()

    device_label = "//h3[text()='Some Device']"
    wait_or_screenshot(driver, ui_mode, screenshot_dir, EC.invisibility_of_element_located((By.XPATH, device_label)))
    by_xpath = driver.find_element_by_xpath(device_label)
    screenshots(driver, screenshot_dir, '{}-{}'.format(screenshot, ui_mode))
    assert by_xpath is not None


def test_account(driver, ui_mode, screenshot_dir):
    menu(driver, ui_mode, screenshot_dir, 'account')

    header_xpath = "//h2[text()='Account']"
    wait_or_screenshot(driver, ui_mode, screenshot_dir, EC.presence_of_element_located((By.XPATH, header_xpath)))

    screenshots(driver, screenshot_dir, 'account-' + ui_mode)


def test_account_notification(driver, ui_mode, screenshot_dir):
    driver.find_element_by_id("chk_email").click()
    driver.find_element_by_id("save").click()
    switch = "//input[@id='chk_email' and @value='false']"
    wait_or_screenshot(driver, ui_mode, screenshot_dir, EC.presence_of_element_located((By.XPATH, switch)))
    screenshots(driver, screenshot_dir, 'account-notification-off-' + ui_mode)

    driver.find_element_by_id("chk_email").click()
    driver.find_element_by_id("save").click()
    switch = "//input[@id='chk_email' and @value='true']"
    wait_or_screenshot(driver, ui_mode, screenshot_dir, EC.presence_of_element_located((By.XPATH, switch)))
    screenshots(driver, screenshot_dir, 'account-notification-on-' + ui_mode)


def test_account_premium_request(driver, ui_mode, screenshot_dir):
    wait_or_screenshot(driver, ui_mode, screenshot_dir, EC.presence_of_element_located((By.ID, 'request_premium')))
    driver.find_element_by_id("request_premium").click()

    confirm_xpath = "//div[@id='premium_confirmation']//button[contains(text(), 'Yes')]"
    wait_or_screenshot(driver, ui_mode, screenshot_dir, EC.presence_of_element_located((By.XPATH, confirm_xpath)))
    driver.find_element_by_xpath(confirm_xpath).click()

    screenshots(driver, screenshot_dir, 'account-premium' + ui_mode)


def test_account_premium_approve(artifact_dir):
    premium_account.premium_approve(DEVICE_USER, artifact_dir)


def test_account_premium_acquire(domain):
    response = api.domain_acquire(domain, 'syncloudexample.com', DEVICE_USER, DEVICE_PASSWORD)
    acquire_response = json.loads(response.text)
    assert acquire_response['success'], response.text
    assert acquire_response['update_token'], response.text


def test_account_premium_delete(driver, ui_mode, screenshot_dir):
    menu(driver, ui_mode, screenshot_dir, 'premium-devices')
    domain_delete(driver, ui_mode, screenshot_dir, 'premium-devices-removed')


def test_account_delete(driver, ui_mode, screenshot_dir):
    wait_or_screenshot(driver, ui_mode, screenshot_dir, EC.presence_of_element_located((By.ID, 'delete')))
    driver.find_element_by_id("delete").click()

    confirm_xpath = "//div[@id='delete_confirmation']//button[contains(text(), 'Yes')]"
    wait_or_screenshot(driver, ui_mode, screenshot_dir, EC.presence_of_element_located((By.XPATH, confirm_xpath)))
    driver.find_element_by_xpath(confirm_xpath).click()

    screenshots(driver, screenshot_dir, 'account-delete' + ui_mode)


def test_teardown(driver):
    driver.quit()


def wait_or_screenshot(driver, ui_mode, screenshot_dir, method):
    wait_driver = WebDriverWait(driver, 30)
    try:
        wait_driver.until(method)
    except Exception as e:
        screenshots(driver, screenshot_dir, 'exception-' + ui_mode)
        raise e


def menu(driver, ui_mode, screenshot_dir, element_id):
    retries = 5
    retry = 0
    while retry < retries:
        try:
            if ui_mode == "mobile":
                navbar = driver.find_element_by_id('navbar')
                navbar.click()
            wait_or_screenshot(driver, ui_mode, screenshot_dir, EC.element_to_be_clickable((By.ID, element_id)))
            screenshots(driver, screenshot_dir, element_id + '-' + ui_mode)
            element = driver.find_element_by_id(element_id)
            element.click()
            if ui_mode == "mobile":
                navbar = driver.find_element_by_id('navbar')
                navbar.click()
            return
        except Exception as e:
            print('error (attempt {0}/{1}): {2}'.format(retry + 1, retries, str(e)))
            time.sleep(1)
        retry += 1

