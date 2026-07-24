import json
import requests


def domain_acquire(hostname, domain, email, password):
    acquire_data = {
        'domain': domain,
        'email': email,
        'password': password,
        'device_mac_address': '00:00:00:00:00:00',
        'device_name': 'some-device',
        'device_title': 'Some Device',
    }
    response = requests.post('https://api.{0}/domain/acquire_v2'.format(hostname),
                             json=acquire_data,
                             verify=False)
    acquire_response = json.loads(response.text)
    assert acquire_response['success'], response.text
    assert acquire_response['data']['update_token'], response.text
    return acquire_response['data']['update_token']


def domain_update(domain, update_token, ip, platform_version='1'):
    update_data = {
        'token': update_token,
        'ip': ip,
        'web_protocol': 'https',
        'web_port': 10001,
        'web_local_port': 80,
        'platform_version': platform_version
    }
    return requests.post('https://api.{0}/domain/update'.format(domain), json=update_data,
                         verify=False)
