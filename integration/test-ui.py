import socket

import time
from os.path import dirname, join
from subprocess import check_output

import pytest
from selenium.webdriver.common.by import By
from selenium.webdriver.common.keys import Keys
from selenium.webdriver.support import expected_conditions as EC
from syncloudlib.integration.hosts import add_host_alias
import smtp

import db
import subscription
import api
import ui

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
        query = 'mysql --host=mysql --user=root --password=root redirect -e'
        check_output("{0} 'select * from action' > {1}/{2}/db-action.log || true".format(query, artifact_dir, ui_mode), shell=True)
        check_output("{0} 'select * from user' > {1}/{2}/db-user.log".format(query, artifact_dir, ui_mode), shell=True)
        check_output("{0} 'select * from domain' > {1}/{2}/db-domain.log".format(query, artifact_dir, ui_mode), shell=True)
        check_output('cp -R {0} {1}'.format(log_dir, artifact_dir), shell=True)
        check_output('cp /videos/* {0}'.format(artifact_dir), shell=True)
        check_output('chmod -R a+r {0}'.format(artifact_dir), shell=True)

    request.addfinalizer(module_teardown)


def test_start(module_setup, device_host, domain):
    add_host_alias('api', device_host, domain)
    db.recreate()


def test_error(domain, selenium):
    ip = socket.gethostbyname('www.syncloud.test')
    print('domain ip: ' + ip)
    selenium.driver.get("https://www.{0}/error".format(domain))
    selenium.find_by(By.ID, 'error')
    selenium.screenshot('error')


def test_index(domain, selenium):
    selenium.driver.get("https://www.{0}".format(domain))
    selenium.find_by(By.ID, 'email')
    selenium.screenshot('index')


def test_register(ui_mode, selenium):
    register(ui_mode, selenium)


def register(ui_mode, selenium):
    menu(selenium, ui_mode, 'register')

    selenium.find_by(By.ID, 'register_email')
    selenium.screenshot('register')
    email = selenium.find_by_id('register_email')
    email.send_keys(DEVICE_USER)
    password = selenium.find_by_id("register_password")
    password.send_keys(DEVICE_PASSWORD)
    selenium.screenshot('register-credentials')
    password.send_keys(Keys.RETURN)
    selenium.find_by(By.XPATH, "//h2[text()='Complete']")

    selenium.screenshot('complete-registration')
    activate_url = smtp.get_activate_url(smtp.emails()[0])
    smtp.clear()
    print('activate_url: ' + activate_url)
    selenium.screenshot('activate-before')
    selenium.driver.get(activate_url)
    selenium.screenshot('activate-after')
    selenium.find_by(By.XPATH, "//span[text()='User was activated']")


def test_login_wrong_username(ui_mode, selenium):
    menu(selenium, ui_mode, 'login')

    selenium.find_by(By.ID, 'email')

    selenium.screenshot('login-wrong-username-' + ui_mode)
    user = selenium.find_by_id("email")
    user.send_keys('wrong_user')
    password = selenium.find_by_id("password")
    password.send_keys('wrong_password')
    selenium.screenshot('login-wrong-username-credentials')
    selenium.find_by_id("submit").click()
    password.send_keys(Keys.RETURN)
    selenium.screenshot('login-wrong-username-progress')
    selenium.find_by(By.ID, 'help-email')
    selenium.screenshot('login-wrong-username-error')
    error = selenium.find_by_id("help-email")
    assert "Not valid email" in error.text


def test_login_wrong_password(ui_mode, selenium):
    menu(selenium, ui_mode, 'login')

    selenium.find_by(By.ID, 'email')

    selenium.screenshot('login-wrong-password-')
    user = selenium.find_by_id("email")
    user.clear()
    user.send_keys('wrong_user@example.com')
    password = selenium.find_by_id("password")
    password.clear()
    password.send_keys('wrong_password')
    selenium.screenshot('login-wrong-password-credentials-')
    password.send_keys(Keys.RETURN)
    selenium.screenshot('login-wrong-password-progress')
    selenium.find_by(By.ID, 'error')
    selenium.screenshot('login-wrong-password-error')
    error = selenium.find_by_id("error")
    assert "authentication failed" in error.text


def test_login(ui_mode, selenium):
    menu(selenium, ui_mode, 'login')

    selenium.find_by(By.ID, 'email')

    selenium.screenshot('login')
    user = selenium.find_by_id("email")
    user.clear()
    user.send_keys(DEVICE_USER)
    password = selenium.find_by_id("password")
    password.clear()
    password.send_keys(DEVICE_PASSWORD)
    selenium.screenshot('login-credentials')
    password.send_keys(Keys.RETURN)
    selenium.screenshot('login-progress')
    selenium.find_by(By.ID, 'no_domains')
    selenium.screenshot('default')
    assert "You do not have any activated devices" in selenium.driver.page_source


def test_devices(domain, ui_mode, artifact_dir, selenium):

    api.domain_acquire(domain, '{}.{}'.format(ui_mode, domain), DEVICE_USER, DEVICE_PASSWORD)

    selenium.driver.get("https://www.{0}/api/domains".format(domain))
    with open(join(artifact_dir, '{}-api-domains.log'.format(ui_mode)), 'w') as f:
        f.write(str(selenium.driver.page_source.encode("utf-8")))
    selenium.driver.get("https://www.{0}".format(domain))

    menu(selenium, ui_mode, 'devices')

    device_label = "//h3[text()='Some Device']"
    selenium.find_by(By.XPATH, device_label)
    by_xpath = selenium.find_by_xpath(device_label)
    selenium.screenshot('devices')
    assert by_xpath is not None


def test_password_reset(ui_mode, selenium):
    menu(selenium, ui_mode, 'logout')

    selenium.wait_or_screenshot(EC.presence_of_element_located((By.ID, 'forgot')))
    forgot = selenium.find_by_id('forgot')
    forgot.click()

    selenium.wait_or_screenshot(EC.presence_of_element_located((By.ID, 'send')))

    email = selenium.find_by_id('email')
    email.send_keys(DEVICE_USER)
    selenium.screenshot('password-forgot')
    send = selenium.find_by_id('send')
    send.click()

    reset_url = smtp.get_reset_url(smtp.emails()[0])
    smtp.clear()
    print('reset_url: ' + reset_url)
    selenium.screenshot('password-reset-before')
    selenium.driver.get(reset_url)
    selenium.screenshot('password-reset-after')
    selenium.find_by(By.ID, 'password')
    password = selenium.find_by(By.ID, 'password')
    global DEVICE_PASSWORD
    DEVICE_PASSWORD = 'password1'
    password.send_keys(DEVICE_PASSWORD)
    reset = selenium.find_by(By.ID, 'reset')
    reset.click()

    menu(selenium, ui_mode, 'login')

    selenium.wait_or_screenshot(EC.presence_of_element_located((By.ID, 'email')))

    selenium.screenshot('reset-login')
    user = selenium.find_by_id("email")
    user.send_keys(DEVICE_USER)
    password = selenium.find_by_id("password")
    password.send_keys(DEVICE_PASSWORD)
    selenium.screenshot('reset-login-credentials')
    password.send_keys(Keys.RETURN)
    selenium.screenshot('reset-login-progress')
    device_label = "//h3[text()='Some Device']"
    selenium.wait_or_screenshot(EC.presence_of_element_located((By.XPATH, device_label)))
    by_xpath = selenium.find_by_xpath(device_label)
    selenium.screenshot('reset-devices')
    assert by_xpath is not None


def test_domain_delete(selenium):
    ui.domain_delete( 'devices-removed', selenium)


def test_account(ui_mode, selenium):
    menu(selenium, ui_mode, 'account')

    header_xpath = "//h2[text()='Account']"
    selenium.wait_or_screenshot(EC.presence_of_element_located((By.XPATH, header_xpath)))

    selenium.screenshot('account')


def test_notification(driver, selenium):
    driver.find_element_by_id("chk_email").click()
    driver.find_element_by_id("save").click()
    switch = "//input[@id='chk_email' and @value='false']"
    selenium.wait_or_screenshot(EC.presence_of_element_located((By.XPATH, switch)))
    selenium.screenshot('account-notification-off')

    driver.find_element_by_id("chk_email").click()
    driver.find_element_by_id("save").click()
    switch = "//input[@id='chk_email' and @value='true']"
    selenium.wait_or_screenshot(EC.presence_of_element_located((By.XPATH, switch)))
    selenium.screenshot('account-notification-on')


def test_subscribe(selenium):
    subscription.subscribe_crypto(selenium)


def test_unsubscribe(selenium):
    selenium.find_by_id('subscription_active')
    selenium.find_by_id('cancel').click()
    selenium.find_by_xpath("//button[contains(., 'Confirm')]").click()
    selenium.find_by_id('subscription_inactive')
    selenium.screenshot('account-subscription-inactive')


def test_resubscribe(selenium):
    subscription.subscribe_crypto(selenium)


def test_domain_when_subscribed(ui_mode, selenium, domain):
    api.domain_acquire(domain, 'syncloudexample.com', DEVICE_USER, DEVICE_PASSWORD)
    menu(selenium, ui_mode, 'devices')
    selenium.screenshot('domain')
    ui.domain_delete('devices-removed-when-subscribed', selenium)


def test_account_delete(driver, ui_mode, selenium):
    menu(selenium, ui_mode, 'account')
    selenium.wait_or_screenshot(EC.presence_of_element_located((By.ID, 'delete')))
    driver.find_element_by_id("delete").click()
    selenium.find_by_xpath("//button[contains(., 'Confirm')]").click()
    selenium.screenshot('account-delete')


def test_user_cleaned_auto_logout(ui_mode, selenium, domain):
    register(ui_mode, selenium)
    sql("delete from user")
    selenium.driver.get("https://www.{0}".format(domain))
    selenium.find_by_id('email')
    selenium.find_by_id("password")
    selenium.screenshot('user-clear')


def sql(query):
    check_output(f"mysql --host=mysql --user=root --password=root redirect -e '{query}'", shell=True)


def test_teardown(driver):
    driver.quit()


def menu(selenium, ui_mode, element_id):
    retries = 5
    retry = 0
    while retry < retries:
        try:
            if ui_mode == "mobile":
                navbar = selenium.find_by_id('navbar')
                navbar.click()
            selenium.wait_or_screenshot(EC.element_to_be_clickable((By.ID, element_id)))
            selenium.screenshot(element_id)
            element = selenium.find_by_id(element_id)
            element.click()
            if ui_mode == "mobile":
                navbar = selenium.find_by_id('navbar')
                navbar.click()
            return
        except Exception as e:
            print('error (attempt {0}/{1}): {2}'.format(retry + 1, retries, str(e)))
            time.sleep(1)
        retry += 1
