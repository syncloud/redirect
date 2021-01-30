import datetime
import json
import quopri
import time
from os.path import join

from urlparse import urlparse

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

    return [quopri.decodestring(message['Content']['Body']) for message in json.loads(response.text)]


def clear():
    response = requests.delete('http://mail:8025/api/v1/messages')
    assert response.status_code == 200, response.text


def get_token(body):
    link_index = body.find('https://')
    link = body[link_index:].split(' ')[0].strip()
    parts = urlparse(link)
    token = parts.query.replace('token=', '')
    return token


def get_activate_url(body):
    return body.split('activate your account: ')[1].strip()


def get_reset_url(body):
    return body.split('reset your password: ')[1].strip()
