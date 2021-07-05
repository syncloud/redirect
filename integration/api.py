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
    return response
