import pytest
import os
from os.path import dirname
from os.path import join, exists
from selenium import webdriver
from selenium.webdriver.common.desired_capabilities import DesiredCapabilities
from selenium.webdriver.firefox.firefox_binary import FirefoxBinary

DIR = dirname(__file__)


def pytest_addoption(parser):
    parser.addoption("--ui-mode", action="store", default="desktop")


@pytest.fixture(scope="session")
def project_dir():
    return join(DIR, '..')


def new_profile(user_agent):
    profile = webdriver.FirefoxProfile()
    profile.set_preference('app.update.auto', False)
    profile.set_preference('app.update.enabled', False)
    profile.set_preference("general.useragent.override", user_agent)
    profile.set_preference("devtools.console.stdout.content", True)

    return profile


def new_driver(profile, log_dir, ui_mode):

    firefox_path = '/tools/firefox/firefox'
    caps = DesiredCapabilities.FIREFOX
    caps["marionette"] = True
    caps['acceptSslCerts'] = True

    binary = FirefoxBinary(firefox_path)

    return webdriver.Firefox(profile, capabilities=caps, log_path="{0}/firefox.{1}.log".format(log_dir, ui_mode),
                             firefox_binary=binary, executable_path='/tools/geckodriver/geckodriver')



@pytest.fixture(scope="module")
def desktop_driver(log_dir, ui_mode):
    profile = new_profile("Mozilla/5.0 (X11; Linux x86_64; rv:10.0) Gecko/20100101 Firefox/10.0")
    driver = new_driver(profile, log_dir, ui_mode)
    driver.set_window_position(0, 0)
    driver.set_window_size(1024, 2000)
    return driver


@pytest.fixture(scope="module")
def mobile_driver(log_dir, ui_mode):
    profile = new_profile("Mozilla/5.0 (iPhone; U; CPU iPhone OS 3_0 like Mac OS X; en-us) AppleWebKit/528.18 (KHTML, like Gecko) Version/4.0 Mobile/7A341 Safari/528.16")
    driver = new_driver(profile, log_dir, ui_mode)
    driver.set_window_position(0, 0)
    driver.set_window_size(400, 2000)
    return driver


@pytest.fixture(scope="module")
def driver(mobile_driver, desktop_driver, ui_mode):
    if ui_mode == "desktop":
        return desktop_driver
    else:
        return mobile_driver

@pytest.fixture(scope='session')
def ui_mode(request):
    return request.config.getoption("--ui-mode")


@pytest.fixture(scope="session")
def log_dir(artifact_dir):
    dir = join(artifact_dir, 'log')
    if not exists(dir):
        os.mkdir(dir)
    return dir


@pytest.fixture(scope="session")
def artifact_dir(project_dir):
    dir =  join(project_dir, 'artifact')
    if not exists(dir):
        os.mkdir(dir)
    return dir


@pytest.fixture(scope="session")
def screenshot_dir(artifact_dir):
    dir = join(artifact_dir, 'screenshot')
    if not exists(dir):
        os.mkdir(dir)
    return dir


