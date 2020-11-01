from os.path import dirname
from subprocess import check_output

import pytest
from selenium.webdriver.common.by import By
from selenium.webdriver.common.keys import Keys
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
from syncloudlib.integration.hosts import add_host_alias_by_ip
from syncloudlib.integration.screenshots import screenshots
import smtp

import db

DIR = dirname(__file__)
DEVICE_USER="user@example.com"
DEVICE_PASSWORD="password"
TMP_DIR = '/tmp/syncloud'


@pytest.fixture(scope="session")
def module_setup(request, ui_mode, log_dir, artifact_dir, device):
    def module_teardown():
        device.run_ssh('ls -la /data/platform/backup > {0}/data.platform.backup.ls.log'.format(TMP_DIR), throw=False)

        device.run_ssh('cp /var/log/apache2/redirect_rest-error.log {0}/{1}-rest-error.log'.format(TMP_DIR, ui_mode), throw=False)
        device.run_ssh('cp /var/log/apache2/redirect_rest-access.log {0}/{1}-rest-access.log'.format(TMP_DIR, ui_mode), throw=False)
        device.run_ssh('cp /var/log/apache2/redirect_ssl_web-access.log {0}/{1}-web-access.log'.format(TMP_DIR, ui_mode), throw=False)
        device.run_ssh('cp /var/log/apache2/redirect_ssl_web-error.log {0}/{1}-web-error.log'.format(TMP_DIR, ui_mode), throw=False)
        device.scp_from_device('{0}/*'.format(TMP_DIR), artifact_dir)
        check_output('cp -R {0} {1}'.format(log_dir, artifact_dir), shell=True)
        check_output('chmod -R a+r {0}'.format(artifact_dir), shell=True)

    request.addfinalizer(module_teardown)


def test_start(module_setup, device_host, domain, driver):
    driver.implicitly_wait(10) 
    check_output('apt-get update', shell=True)
    check_output('apt-get install -y mysql-client', shell=True)
    add_host_alias_by_ip('app', 'www', device_host, domain)
    add_host_alias_by_ip('app', 'api', device_host, domain)
    db.recreate()


def test_error(driver, screenshot_dir, ui_mode, domain):
    driver.get("https://www.{0}/error".format(domain))
    wait_or_screenshot(driver, ui_mode, screenshot_dir, EC.presence_of_element_located((By.ID, 'error')))
    screenshots(driver, screenshot_dir, 'error-' + ui_mode)


def test_index(driver, screenshot_dir, ui_mode, domain):
    driver.get("https://www.{0}".format(domain))
    wait_or_screenshot(driver, ui_mode, screenshot_dir, EC.presence_of_element_located((By.ID, 'email')))
    screenshots(driver, screenshot_dir, 'index-' + ui_mode)


def test_register(driver, ui_mode, screenshot_dir):
    menu(driver, ui_mode)
    register = driver.find_element_by_id('register')
    register.click()
    menu(driver, ui_mode)

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
    menu(driver, ui_mode)
    wait_or_screenshot(driver, ui_mode, screenshot_dir, EC.presence_of_element_located((By.ID, 'login')))
    login = driver.find_element_by_id('login')
    login.click()
    menu(driver, ui_mode)

    wait_or_screenshot(driver, ui_mode, screenshot_dir, EC.presence_of_element_located((By.ID, 'email')))

    screenshots(driver, screenshot_dir, 'login-' + ui_mode)
    user = driver.find_element_by_id("email")
    user.send_keys(DEVICE_USER)
    password = driver.find_element_by_id("password")
    password.send_keys(DEVICE_PASSWORD)
    screenshots(driver, screenshot_dir, 'login-credentials-' + ui_mode)
    password.send_keys(Keys.RETURN)
    screenshots(driver, screenshot_dir, 'login-progress-' + ui_mode)
    wait_or_screenshot(driver, ui_mode, screenshot_dir, EC.presence_of_element_located((By.ID, 'no_domains')))
    screenshots(driver, screenshot_dir, 'default-' + ui_mode)
    assert "You do not have any activated devices" in driver.page_source.encode("utf-8")


def test_devices(driver, ui_mode, screenshot_dir):
    menu(driver, ui_mode)
    wait_or_screenshot(driver, ui_mode, screenshot_dir, EC.presence_of_element_located((By.ID, 'devices')))
    devices = driver.find_element_by_id('devices')
    devices.click()
    menu(driver, ui_mode)

    wait_or_screenshot(driver, ui_mode, screenshot_dir, EC.presence_of_element_located((By.ID, 'no_domains')))
    screenshots(driver, screenshot_dir, 'devices-' + ui_mode)
    assert "You do not have any activated devices" in driver.page_source.encode("utf-8")


def test_password_reset(driver, ui_mode, screenshot_dir):
    menu(driver, ui_mode)
    wait_or_screenshot(driver, ui_mode, screenshot_dir, EC.presence_of_element_located((By.ID, 'logout')))
    screenshots(driver, screenshot_dir, 'logout-' + ui_mode)
    logout = driver.find_element_by_id('logout')
    logout.click()
    menu(driver, ui_mode)

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

    test_login(driver, ui_mode, screenshot_dir)


def test_account(driver, ui_mode, screenshot_dir):
    menu(driver, ui_mode)
    wait_or_screenshot(driver, ui_mode, screenshot_dir, EC.presence_of_element_located((By.ID, 'account')))
    account = driver.find_element_by_id('account')
    account.click()
    menu(driver, ui_mode)

    delete_btn_xpath = "//button[text()='Delete']"
    wait_or_screenshot(driver, ui_mode, screenshot_dir, EC.presence_of_element_located((By.XPATH, delete_btn_xpath)))

    screenshots(driver, screenshot_dir, 'account-' + ui_mode)


def wait_or_screenshot(driver, ui_mode, screenshot_dir, method):
    wait_driver = WebDriverWait(driver, 10)
    try:
        wait_driver.until(method)
    except Exception as e:
        screenshots(driver, screenshot_dir, 'exception-' + ui_mode)
        raise e


def menu(driver, ui_mode):
    if ui_mode == "mobile":
        navbar = driver.find_element_by_id('navbar')
        navbar.click()
