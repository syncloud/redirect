import time
from os import environ
from os.path import dirname
from subprocess import check_output
from os.path import join
import json
import pytest
import requests
from syncloudlib.integration.hosts import add_host_alias_by_ip

import db
import smtp

DIR = dirname(__file__)
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
    
        device.scp_from_device('{0}/*'.format(TMP_DIR), artifact_dir)
        check_output('chmod -R a+r {0}'.format(artifact_dir), shell=True)
        db.recreate()

    request.addfinalizer(module_teardown)


def test_start(module_setup, device, device_host, domain, build_number):

    check_output('apt-get update', shell=True)
    check_output('apt-get install -y mysql-client', shell=True)

    add_host_alias_by_ip('app', 'www', device_host, domain)
    add_host_alias_by_ip('app', 'api', device_host, domain)
    device.run_ssh('mkdir {0}'.format(TMP_DIR))
    device.run_ssh("snap remove platform")
    device.run_ssh("apt-get update")
    device.run_ssh("apt-get install -y mysql-client default-libmysqlclient-dev apache2 python libpython2.7 python-pip libapache2-mod-wsgi python-mysqldb python-dev openssl > {0}/apt.log".format(TMP_DIR))
    device.scp_to_device("fakecertificate.sh", "/")
    device.run_ssh("/fakecertificate.sh")
    device.scp_to_device("../artifact/redirect-{0}.tar.gz".format(build_number), "/")
    device.scp_to_device("../ci/deploy", "/")
    device.run_ssh("cd / && /deploy {0} integration syncloud.test > {1}/deploy.log 2>&1".format(build_number, TMP_DIR))
    device.run_ssh("sed -i 's#@access_key_id@#{0}#g' /var/www/redirect/secret.cfg".format(environ['access_key_id']), debug=False)
    device.run_ssh("sed -i 's#@secret_access_key@#{0}#g' /var/www/redirect/secret.cfg".format(environ['secret_access_key']), debug=False)
    device.run_ssh("sed -i 's#@hosted_zone_id@#{0}#g' /var/www/redirect/secret.cfg".format(environ['hosted_zone_id']), debug=False)
    device.run_ssh("systemctl restart redirect")


def get_domain(update_token, domain):
    response = requests.get('https://api.{0}/domain/get'.format(domain),
                            params={'token': update_token}, verify=False)
    assert response.status_code == 200
    assert response.text is not None
    response_data = json.loads(response.text)
    return response_data['data']


def test_index(domain, artifact_dir):
    response = requests.get('https://www.{0}'.format(domain), allow_redirects=False, verify=False)
    assert response.status_code == 200, response.text
    with open(join(artifact_dir, 'index.html.log'), 'w') as f:
       f.write(str(response.text))


def test_user_create_special_symbols_in_password(domain):
    email = 'symbols_in_password@mail.com'
    response = requests.post('https://www.{0}/api/user/create'.format(domain),
                             data={'email': email, 'password': r'pass12& ^%"'},
                             verify=False)
    assert response.status_code == 200, response.text
    assert len(smtp.emails()) == 1
    smtp.clear()


def create_user(domain, email, password):
    response = requests.post('https://www.{0}/api/user/create'.format(domain),
                             data={'email': email, 'password': password}, verify=False)
    assert response.status_code == 200, response.text

    activate_user(domain)

    response = requests.get('https://api.{0}/user/get'.format(domain),
                            params={'email': email, 'password': password},
                            verify=False)
    assert response.status_code == 200, response.text


def test_create_user_api_for_mobile_app(domain):
    email = 'mobile_create_user@syncloud.test'
    password = 'pass123456'
    response = requests.post('https://api.{0}/user/create'.format(domain),
                             data={'email': email, 'password': password}, verify=False)
    assert response.status_code == 200, response.text

    activate_user(domain)

    response = requests.get('https://api.{0}/user/get'.format(domain),
                            params={'email': email, 'password': password},
                            verify=False)
    assert response.status_code == 200, response.text


def activate_user(domain):
    assert len(smtp.emails()) == 1
    activate_token = smtp.get_token(smtp.emails()[0])
    response = requests.post('https://www.{0}/api/user/activate'.format(domain),
                             data={'token': activate_token},
                             verify=False)
    assert response.status_code == 200, response.text
    smtp.clear()


def acquire_domain(domain, email, password, user_domain):
    acquire_data = {
        'user_domain': user_domain,
        'email': email,
        'password': password,
        'device_mac_address': '00:00:00:00:00:00',
        'device_name': 'some-device',
        'device_title': 'Some Device',
    }
    response = requests.post('https://api.{0}/domain/acquire'.format(domain),
                             data=acquire_data,
                             verify=False)
    domain_data = json.loads(response.text)
    assert 'update_token' in domain_data, response.text
    update_token = domain_data['update_token']
    return update_token


def test_user_create_success(domain):
    create_user(domain, 'test@syncloud.test', 'pass123456')


def test_user_create_existing_case_difference(domain):
    email1 = 'case_test@syncloud.test'
    email2 = 'Case_test@syncloud.test'
    create_user(domain, email1, 'pass123456')
    response = requests.post('https://www.{0}/api/user/create'.format(domain),
                             data={'email': email2, 'password': 'pass123456'}, verify=False)
    assert response.status_code == 400, response.text
    assert "already registered" in response.text, response.text


def test_get_user_data(domain):
    email = 'test_get_user_data@syncloud.test'
    password = 'pass123456'
    create_user(domain, email, password)

    user_domain = "test_get_user_data"
    update_token = acquire_domain(domain, email, password, user_domain)

    update_data = {
        'token': update_token,
        'ip': '192.192.1.1',
        'web_protocol': 'http',
        'web_local_port': 80,
        'web_port': 10000
    }

    response = requests.post('https://api.{0}/domain/update'.format(domain),
                  json=update_data,
                  verify=False)
    assert response.status_code == 200, response.text

    response_data = json.loads(response.text)
    assert response_data['success'], response.text

    response = requests.get('https://api.{0}/user/get'.format(domain),
                            params={'email': email, 'password': password},
                            verify=False)

    response_data = json.loads(response.text)
    user_data = response_data['data']

    # This is hack. We do not know last_update value - it is set by server.
    last_update = user_data["domains"][0]["last_update"]
    update_token = user_data["update_token"]

    expected = {
        'active': True,
        'email': email,
        'unsubscribed': False,
        'update_token': update_token,
        'domains': [{
            'user_domain': user_domain,
            'web_local_port': 80,
            'web_port': 10000,
            'web_protocol': 'http',
            'ip': '192.192.1.1',
            'ipv6': None,
            'dkim_key': None,
            'local_ip': None,
            'map_local_address': False,
            'platform_version': None,
            'device_mac_address': '00:00:00:00:00:00',
            'device_name': 'some-device',
            'device_title': 'Some Device',
            'last_update': last_update
        }]
    }

    assert expected == user_data


def test_user_delete(domain):
    email = 'test_user_delete@syncloud.test'
    password = 'pass123456'
    create_user(domain, email, password)

    update_token_1 = acquire_domain(domain, email, password, "user_domain_1")
    update_token_2 = acquire_domain(domain, email, password, "user_domain_2")

    response = requests.post('https://api.{0}/user/delete'.format(domain),
                             data={'email': email, 'password': password}, verify=False)
    assert response.status_code == 200

    response = requests.get('https://api.{0}/domain/get'.format(domain),
                            params={'token': update_token_1}, verify=False)
    assert response.status_code == 400

    response = requests.get('https://api.{0}/domain/get'.format(domain),
                            params={'token': update_token_2},
                            verify=False)
    assert response.status_code == 400


def test_user_reset_password_sent_mail(domain):
    email = 'test_user_reset_password_sent_mail@syncloud.test'
    password = 'pass123456'
    create_user(domain, email, password)

    response = requests.post('https://www.{0}/api/user/reset_password'.format(domain),
                             data={'email': email}, verify=False)
    assert response.status_code == 200

    assert len(smtp.emails()) > 0, 'Server should send email with link to reset password'
    token = smtp.get_token(smtp.emails()[0])
    smtp.clear()
    assert token is not None


def test_user_reset_password_set_new(domain):
    email = 'test_user_reset_password_set_new@syncloud.test'
    password = 'pass123456'
    create_user(domain, email, password)

    requests.post('https://www.{0}/api/user/reset_password'.format(domain), data={'email': email},
                  verify=False)
    token = smtp.get_token(smtp.emails()[0])

    smtp.clear()

    new_password = 'new_password'
    response = requests.post('https://www.{0}/api/user/set_password'.format(domain),
                             data={'token': token, 'password': new_password},
                             verify=False)
    assert response.status_code == 200, response.text

    assert len(smtp.emails()) > 0, 'Server should send email when setting new password'

    response = requests.get('https://api.{0}/user/get'.format(domain),
                            params={'email': email, 'password': new_password},
                            verify=False)
    assert response.status_code == 200, response.text
    smtp.clear()


def test_user_reset_password_set_with_old_token(domain):
    email = 'test_user_reset_password_set_with_old_token@syncloud.test'
    password = 'pass123456'
    create_user(domain, email, password)

    requests.post('https://www.{0}/api/user/reset_password'.format(domain), data={'email': email},
                  verify=False)
    token_old = smtp.get_token(smtp.emails()[0])

    smtp.clear()

    requests.post('https://www.{0}/api/user/reset_password'.format(domain), data={'email': email},
                  verify=False)
    token = smtp.get_token(smtp.emails()[0])
    smtp.clear()

    new_password = 'new_password'
    response = requests.post('https://www.{0}/api/user/set_password'.format(domain),
                             data={'token': token_old, 'password': new_password},
                             verify=False)
    assert response.status_code == 400, response.text
    smtp.clear()


def test_user_reset_password_set_twice(domain):
    email = 'test_user_reset_password_set_twice@syncloud.test'
    password = 'pass123456'
    create_user(domain, email, password)

    requests.post('https://www.{0}/api/user/reset_password'.format(domain), data={'email': email},
                  verify=False)
    token = smtp.get_token(smtp.emails()[0])
    smtp.clear()

    new_password = 'new_password'
    response = requests.post('https://www.{0}/api/user/set_password'.format(domain),
                             data={'token': token, 'password': new_password},
                             verify=False)
    assert response.status_code == 200, response.text

    new_password = 'new_password2'
    response = requests.post('https://www.{0}/api/user/set_password'.format(domain),
                             data={'token': token, 'password': new_password},
                             verify=False)
    assert response.status_code == 400, response.text
    smtp.clear()


def test_domain_new(domain):
    email = 'test_domain_new@syncloud.test'
    password = 'pass123456'
    create_user(domain, email, password)

    user_domain = "test_domain_new"
    acquire_data = dict(
        user_domain=user_domain,
        device_mac_address='00:00:00:00:00:00',
        device_name='my-super-board',
        device_title='My Super Board',
        email=email,
        password=password)
    response = requests.post('https://api.{0}/domain/acquire'.format(domain), data=acquire_data,
                             verify=False)

    assert response.status_code == 200
    domain_data = json.loads(response.text)

    update_token = domain_data['update_token']
    assert update_token is not None

    expected_data = {
        'update_token': update_token,
        'user_domain': user_domain,
        'device_mac_address': '00:00:00:00:00:00',
        'device_name': 'my-super-board',
        'device_title': 'My Super Board'
    }

    data = get_domain(update_token, domain)
    data.pop('last_update', None)
    assert expected_data == data


def test_domain_new_v2(domain):
    email = 'test_domain_new_v2@syncloud.test'
    password = 'pass123456'
    create_user(domain, email, password)

    user_domain = "test_domain_new_v2"
    acquire_data = dict(
        user_domain=user_domain,
        device_mac_address='00:00:00:00:00:00',
        device_name='my-super-board',
        device_title='My Super Board',
        email=email,
        password=password)
    response = requests.post('https://api.{0}/domain/acquire_v2'.format(domain), json=acquire_data,
                             verify=False)

    assert response.status_code == 200
    domain_data = json.loads(response.text)['data']

    update_token = domain_data['update_token']
    assert update_token is not None

    expected_data = {
        'update_token': update_token,
        'user_domain': user_domain,
        'device_mac_address': '00:00:00:00:00:00',
        'device_name': 'my-super-board',
        'device_title': 'My Super Board'
    }

    data = get_domain(update_token, domain)
    data.pop('last_update', None)
    assert expected_data == data


def test_domain_existing(domain):
    email_1 = 'test_domain_existing_@syncloud.test'
    password_1 = 'pass123456_'
    create_user(domain, email_1, password_1)

    user_domain = "test_domain_existing"
    acquire_data = dict(
        user_domain=user_domain,
        device_mac_address='00:00:00:00:00:00',
        device_name='my-super-board',
        device_title='My Super Board',
        email=email_1,
        password=password_1)
    response = requests.post('https://api.{0}/domain/acquire'.format(domain), data=acquire_data,
                             verify=False)
    domain_data = json.loads(response.text)
    update_token = domain_data['update_token']

    email_2 = 'test_domain_existing@syncloud.test'
    password_2 = 'pass123456'
    create_user(domain, email_2, password_2)
    acquire_data = dict(
        user_domain=user_domain,
        device_mac_address='00:00:00:00:00:11',
        device_name='other-board',
        device_title='Other Board',
        email=email_2,
        password=password_2)
    response = requests.post('https://api.{0}/domain/acquire'.format(domain), data=acquire_data,
                             verify=False)

    assert response.status_code == 400

    expected_data = {
        'update_token': update_token,
        'user_domain': user_domain,
        'device_mac_address': '00:00:00:00:00:00',
        'device_name': 'my-super-board',
        'device_title': 'My Super Board'
    }

    data = get_domain(update_token, domain)
    data.pop('last_update', None)
    assert expected_data == data


def test_domain_twice(domain):
    email = 'test_domain_twice@syncloud.test'
    password = 'pass123456_'
    create_user(domain, email, password)

    user_domain = "test_domain_twice"
    acquire_data = dict(
        user_domain=user_domain,
        device_mac_address='00:00:00:00:00:00',
        device_name='my-super-board',
        device_title='My Super Board',
        email=email,
        password=password)
    response = requests.post('https://api.{0}/domain/acquire'.format(domain), data=acquire_data,
                             verify=False)
    domain_data = json.loads(response.text)
    update_token1 = domain_data['update_token']

    acquire_data = dict(
        user_domain=user_domain,
        device_mac_address='00:00:00:00:00:11',
        device_name='my-super-board-2',
        device_title='My Super Board 2',
        email=email,
        password=password)
    response = requests.post('https://api.{0}/domain/acquire'.format(domain), data=acquire_data,
                             verify=False)
    assert response.status_code == 200
    domain_data = json.loads(response.text)
    update_token2 = domain_data['update_token']

    assert update_token1 != update_token2

    expected_data = {
        'user_domain': user_domain,
        'update_token': update_token2,
        'device_mac_address': '00:00:00:00:00:11',
        'device_name': 'my-super-board-2',
        'device_title': 'My Super Board 2'
    }

    data = get_domain(update_token2, domain)
    data.pop('last_update', None)
    assert expected_data == data

def test_domain_wrong_mac_address_format(domain):
    email = 'test_domain_wrong_mac_address_format@syncloud.test'
    password = 'pass123456_'
    create_user(domain, email, password)

    user_domain = "test_domain_wrong_mac_address_format"
    acquire_data = {
        'user_domain': user_domain,
        'device_mac_address': 'wrong_format',
        'device_name': 'my-super-board',
        'device_title': 'My Super Board',
        'email': email,
        'password': password
    }
    response = requests.post('https://api.{0}/domain/acquire'.format(domain), data=acquire_data,
                             verify=False)

    assert response.status_code == 400


def test_drop_device(domain):
    email = 'test_drop_device@syncloud.test'
    password = 'pass123456_'
    create_user(domain, email, password)

    user_domain = "test_drop_device"
    acquire_data = dict(
        user_domain=user_domain,
        device_mac_address='00:00:00:00:00:00',
        device_name='my-super-board',
        device_title='My Super Board',
        email=email,
        password=password)
    response = requests.post('https://api.{0}/domain/acquire'.format(domain), data=acquire_data,
                             verify=False)

    assert response.status_code == 200
    domain_data = json.loads(response.text)

    update_token = domain_data['update_token']

    drop_data = {
        'email': email,
        'password': password,
        'user_domain': user_domain
    }

    response = requests.post('https://api.{0}/domain/drop_device'.format(domain), data=drop_data,
                             verify=False)
    assert response.status_code == 200

    response = requests.get('https://api.{0}/domain/get'.format(domain), params={'token': update_token},
                            verify=False)
    assert response.status_code == 400


def test_domain_delete(domain):
    email = 'test_domain_delete@syncloud.test'
    password = 'pass123456'
    create_user(domain, email, password)

    user_domain = "test_domain_delete"
    update_token = acquire_domain(domain, email, password, user_domain)

    delete_data = {'user_domain': user_domain, 'email': email, 'password': password}

    response = requests.post('https://api.{0}/domain/delete'.format(domain), data=json.dumps(delete_data),
                             verify=False)
    assert response.status_code == 200, response.text

    response = requests.get('https://api.{0}/user/get'.format(domain),
                            params={'email': email, 'password': password},
                            verify=False)
    assert response.status_code == 200, response.text

    response_data = json.loads(response.text)
    assert len(response_data['data']['domains']) == 0


def test_domain_update_date(domain):
    email = 'test_domain_update_date@syncloud.test'
    password = 'pass123456'
    create_user(domain, email, password)

    user_domain = "test_domain_update_date"

    update_token = acquire_domain(domain, email, password, user_domain)

    update_data = {
        'token': update_token,
        'ip': '127.0.0.1',
        'web_protocol': 'http',
        'web_port': 10001,
        'web_local_port': 80
    }

    requests.post('https://api.{0}/domain/update'.format(domain), json=update_data,
                  verify=False)
    domain_info = get_domain(update_token, domain)
    last_updated1 = domain_info['last_update']

    time.sleep(1)

    requests.post('https://api.{0}/domain/update'.format(domain), json=update_data,
                  verify=False)
    domain_info = get_domain(update_token, domain)
    last_updated2 = domain_info['last_update']

    assert last_updated2 > last_updated1


def test_domain_update_wrong_token(domain):
    update_data = {'token': 'test_domain_update_wrong_token', 'ip': '127.0.0.1'}

    response = requests.post('https://api.{0}/domain/update'.format(domain), json=update_data,
                             verify=False)
    assert response.status_code == 400, response.text


def test_domain_update_web_updated(domain):
    email = 'test_domain_update_web_updated@syncloud.test'
    password = 'pass123456'
    create_user(domain, email, password)

    user_domain = "test_domain_update_web_updated"
    update_token = acquire_domain(domain, email, password, user_domain)

    update_data = {
        'token': update_token,
        'ip': '127.0.0.1',
        'web_protocol': 'http',
        'web_port': 10001,
        'web_local_port': 80,
    }

    response = requests.post('https://api.{0}/domain/update'.format(domain), json=update_data,
                             verify=False)
    assert response.status_code == 200

    update_data = {
        'token': update_token,
        'ip': '127.0.0.1',
        'web_protocol': 'https',
        'web_port': 10002,
        'web_local_port': 443,
    }

    response = requests.post('https://api.{0}/domain/update'.format(domain), json=update_data,
                             verify=False)

    assert response.status_code == 200

    expected_data = {
        'update_token': update_token,
        'ip': '127.0.0.1',
        'user_domain': user_domain,
        'web_protocol': 'https',
        'web_port': 10002,
        'web_local_port': 443,
        'device_mac_address': '00:00:00:00:00:00',
        'device_name': 'some-device',
        'device_title': 'Some Device'
    }

    domain_data = get_domain(update_token, domain)
    domain_data.pop('last_update', None)
    assert expected_data == domain_data


def test_domain_update_ip_changed(domain):
    email = 'test_domain_update_ip_changed@syncloud.test'
    password = 'pass123456'
    create_user(domain, email, password)
    user_domain = "test_domain_update_ip_changed"
    update_token = acquire_domain(domain, email, password, user_domain)

    update_data = {
        'token': update_token,
        'ip': '127.0.0.1',
        'web_protocol': 'http',
        'web_port': 10001,
        'web_local_port': 80,
    }

    response = requests.post('https://api.{0}/domain/update'.format(domain), json=update_data,
                             verify=False)
    assert response.status_code == 200

    update_data = {
        'token': update_token,
        'ip': '127.0.0.2',
        'web_protocol': 'http',
        'web_port': 10001,
        'web_local_port': 80,
    }

    response = requests.post('https://api.{0}/domain/update'.format(domain), json=update_data,
                             verify=False)

    assert response.status_code == 200

    expected_data = {
        'update_token': update_token,
        'ip': '127.0.0.2',
        'user_domain': user_domain,
        'device_mac_address': '00:00:00:00:00:00',
        'device_name': 'some-device',
        'device_title': 'Some Device',
        'web_local_port': 80,
        'web_port': 10001,
        'web_protocol': 'http'
    }

    domain_data = get_domain(update_token, domain)
    domain_data.pop('last_update', None)
    assert expected_data == domain_data

def test_domain_update_platform_version(domain):
    email = 'test_domain_update_platform_version@syncloud.test'
    password = 'pass123456'
    create_user(domain, email, password)
    user_domain = "test_domain_update_platform_version"

    update_token = acquire_domain(domain, email, password, user_domain)

    update_data = {
        'token': update_token,
        'ip': '127.0.0.1',
        'platform_version': '366',
        'web_protocol': 'http',
        'web_port': 10001,
        'web_local_port': 80,
    }

    response = requests.post('https://api.{0}/domain/update'.format(domain), json=update_data,
                             verify=False)
    assert response.status_code == 200

    expected_data = {
        'update_token': update_token,
        'platform_version': '366',
        'device_mac_address': '00:00:00:00:00:00',
        'device_name': 'some-device',
        'device_title': 'Some Device',
        'ip': '127.0.0.1',
        'user_domain': 'test_domain_update_platform_version',
        'web_local_port': 80,
        'web_port': 10001,
        'web_protocol': 'http'
    }
    domain_data = get_domain(update_token, domain)
    domain_data.pop('last_update', None)
    assert expected_data == domain_data


def test_domain_update_local_ip_changed(domain):
    email = 'test_domain_update_local_ip_changed@syncloud.test'
    password = 'pass123456'
    create_user(domain, email, password)
    user_domain = "test_domain_update_local_ip_changed"

    update_token = acquire_domain(domain, email, password, user_domain)

    update_data = {
        'token': update_token,
        'ip': '127.0.0.1',
        'local_ip': '192.168.1.5',
        'web_protocol': 'http',
        'web_port': 10001,
        'web_local_port': 80,
    }

    response = requests.post('https://api.{0}/domain/update'.format(domain), json=update_data,
                             verify=False)
    assert response.status_code == 200

    update_data = {
        'token': update_token,
        'ip': '127.0.0.2',
        'local_ip': '192.168.1.6',
        'web_protocol': 'http',
        'web_port': 10001,
        'web_local_port': 80,
    }

    response = requests.post('https://api.{0}/domain/update'.format(domain), json=update_data,
                             verify=False)

    assert response.status_code == 200

    expected_data = {
        'update_token': update_token,
        'ip': '127.0.0.2',
        'local_ip': '192.168.1.6',
        'user_domain': user_domain,
        'device_mac_address': '00:00:00:00:00:00',
        'device_name': 'some-device',
        'device_title': 'Some Device',
        'web_local_port': 80,
        'web_port': 10001,
        'web_protocol': 'http'
    }
    domain_data = get_domain(update_token, domain)
    domain_data.pop('last_update', None)
    assert expected_data == domain_data


def test_domain_update_server_side_client_ip(domain):
    email = 'test_domain_update_server_side_client_ip@syncloud.test'
    password = 'pass123456'
    create_user(domain, email, password)
    user_domain = "test_domain_update_server_side_client_ip"

    update_token = acquire_domain(domain, email, password, user_domain)

    update_data = {
        'token': update_token,
        'web_protocol': 'http',
        'web_port': 10001,
        'web_local_port': 80,
    }

    response = requests.post('https://api.{0}/domain/update'.format(domain),
                             json=update_data,
                             verify=False)
    assert response.status_code == 200, response.text

    expected_data = {
        'update_token': update_token,
        'user_domain': user_domain,
        'web_protocol': 'http',
        'web_port': 10001,
        'web_local_port': 80,
        'device_mac_address': '00:00:00:00:00:00',
        'device_name': 'some-device',
        'device_title': 'Some Device'
    }

    domain_data = get_domain(update_token, domain)
    domain_data.pop('last_update', None)
    domain_data.pop('ip', None)
    assert expected_data == domain_data


def test_domain_update_map_local_address(domain):
    email = 'test_domain_update_map_local_address@syncloud.test'
    password = 'pass123456'
    create_user(domain, email, password)

    user_domain = "test_domain_update_map_local_address"
    update_token = acquire_domain(domain, email, password, user_domain)

    update_data = {
        'token': update_token,
        'ip': '108.108.108.108',
        'local_ip': '192.168.1.2',
        'map_local_address': True,
        'web_protocol': 'http',
        'web_port': 10001,
        'web_local_port': 80
    }

    response = requests.post('https://api.{0}/domain/update'.format(domain), json=update_data,
                             verify=False)
    assert response.status_code == 200

    expected_data = {
        'update_token': update_token,
        'ip': '108.108.108.108',
        'local_ip': '192.168.1.2',
        'map_local_address': True,
        'user_domain': user_domain,
        'device_mac_address': '00:00:00:00:00:00',
        'device_name': 'some-device',
        'device_title': 'Some Device',
        'web_protocol': 'http',
        'web_port': 10001,
        'web_local_port': 80
    }

    domain_data = get_domain(update_token, domain)
    domain_data.pop('last_update', None)
    assert expected_data == domain_data


def test_status(domain):
    response = requests.get('https://api.{0}/status'.format(domain), verify=False)
    assert response.status_code == 200
    assert 'OK' in response.text


def test_backup(device):
    device.run_ssh("/var/www/redirect/current/bin/redirectdb backup redirect redirect.sql")

