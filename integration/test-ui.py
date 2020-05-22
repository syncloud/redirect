import time
from os.path import dirname
from subprocess import check_output

import pytest
from selenium.webdriver.common.by import By
from selenium.webdriver.common.keys import Keys
from selenium.webdriver.support import expected_conditions as EC
from selenium.webdriver.support.ui import WebDriverWait
from syncloudlib.integration.hosts import add_host_alias_by_ip
from syncloudlib.integration.screenshots import screenshots

DIR = dirname(__file__)
DEVICE_USER="user"
DEVICE_PASSWORD="password"

@pytest.fixture(scope="session")
def module_setup(request, ui_mode, log_dir, artifact_dir):
    def module_teardown():
        check_output('cp /var/log/apache2/error.log {0}/apache.ui.{1}.error.log'.format(log_dir, ui_mode), shell=True)
        check_output('chmod -R a+r {0}'.format(artifact_dir), shell=True)

    request.addfinalizer(module_teardown)


def test_start(module_setup, domain):
    add_host_alias_by_ip('app', 'www', '127.0.0.1', domain)


def test_login(driver, screenshot_dir, ui_mode, domain):
    driver.get("https://www.{0}".format(domain))
    screenshots(driver, screenshot_dir, 'index-' + ui_mode)
    time.sleep(10)


def test_register(driver, ui_mode, screenshot_dir, domain):
    driver.get("https://www.{0}/register.html".format(domain))
    screenshots(driver, screenshot_dir, 'register-' + ui_mode)


def test_main(driver, ui_mode, screenshot_dir, domain):
    driver.get("https://www.{0}".format(domain))
    user = driver.find_element_by_id("user")
    user.send_keys(DEVICE_USER)
    password = driver.find_element_by_id("password")
    password.send_keys(DEVICE_PASSWORD)
    screenshots(driver, screenshot_dir, 'login-' + ui_mode)

    password.send_keys(Keys.RETURN)
    time.sleep(10)
    screenshots(driver, screenshot_dir, 'login_progress-' + ui_mode)
       
    wait_driver = WebDriverWait(driver, 300)

    if ui_mode == "desktop":
        close_btn_xpath =  "//button[@aria-label='Close']"
        wait_driver.until(EC.presence_of_element_located((By.XPATH, close_btn_xpath)))
        wizard_close_button = driver.find_element_by_xpath(close_btn_xpath)
        screenshots(driver, screenshot_dir, 'main_first_time-' + ui_mode)
        wizard_close_button.click()
    
    time.sleep(2)
    screenshots(driver, screenshot_dir, 'main-' + ui_mode)
