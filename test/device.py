import time

import requests


class Device:

    def __init__(self, host):
        self.url = 'http://{0}:4580/faker'.format(host)

    def reset(self):
        response = requests.post('{0}/reset'.format(self.url), timeout=60)
        assert response.status_code == 200, response.text

    def tunnel(self, domain_name, update_token):
        response = requests.post('{0}/tunnel'.format(self.url),
                                 json={'domain': domain_name, 'token': update_token},
                                 timeout=60)
        assert response.status_code == 200, response.text

    def behaviour(self, rcpt='accept', data='accept'):
        response = requests.post('{0}/behaviour'.format(self.url),
                                 json={'rcpt': rcpt, 'data': data}, timeout=30)
        assert response.status_code == 200, response.text

    def messages(self, expected=1, attempts=30):
        messages = []
        for _ in range(attempts):
            response = requests.get('{0}/messages'.format(self.url), timeout=30)
            assert response.status_code == 200, response.text
            messages = response.json()
            if len(messages) >= expected:
                return messages
            time.sleep(1)
        return messages
