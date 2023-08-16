import datetime
import json
import quopri
import time
from os.path import join
import re
from urllib.parse import urlparse

import requests


def emails(artifact_dir=None):
    attempts = 1
    while attempts < 10:
        results = try_emails(artifact_dir)
        if len(results) > 0:
            return results
        attempts += 1
        time.sleep(1)
    return []


def try_emails(artifact_dir):
    response = requests.get('http://mail:8025/api/v1/messages')
    assert response.status_code == 200, response.text
    if artifact_dir:
        with open(join(artifact_dir, 'mails-{}.log'.format(datetime.datetime.now().microsecond)), 'w') as f:
            f.write(str(response.text))

    return [quopri.decodestring(message['Content']['Body']).decode("utf-8") for message in json.loads(response.text)]


def clear():
    response = requests.delete('http://mail:8025/api/v1/messages')
    assert response.status_code == 200, response.text


def get_token(body):
    return re.search(r'https://.*token=(.*)\r', body.replace('=\r\n', '')).group(1)


def get_activate_url(body):
    return re.search(r'activate your account: (https://.*)\r', body.replace('=\r\n', '')).group(1)


def get_reset_url(body):
    return re.search(r'reset your password: (https://.*)\r', body.replace('=\r\n', '')).group(1)
