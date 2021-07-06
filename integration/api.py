import requests
import json


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
    return response
