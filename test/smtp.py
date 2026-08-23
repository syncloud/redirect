import datetime
import json
import quopri
import time
from os.path import join
import re

import requests


def emails(artifact_dir=None, timeout=30, poll=0.25):
    deadline = time.monotonic() + timeout
    while True:
        results = try_emails(None)
        if len(results) > 0:
            if artifact_dir:
                try_emails(artifact_dir)
            return results
        if time.monotonic() >= deadline:
            if artifact_dir:
                try_emails(artifact_dir)
            return []
        time.sleep(poll)


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


def token_of(body):
    match = re.search(r'https://.*token=(.*)\r', body.replace('=\r\n', ''))
    return match.group(1) if match else None


def get_token(body):
    token = token_of(body)
    assert token is not None, 'email carries no token'
    return token


def wait_for(match, artifact_dir=None, timeout=30, poll=0.25):
    deadline = time.monotonic() + timeout
    while True:
        for body in try_emails(None):
            value = match(body)
            if value is not None:
                if artifact_dir:
                    try_emails(artifact_dir)
                return value
        if time.monotonic() >= deadline:
            if artifact_dir:
                try_emails(artifact_dir)
            return None
        time.sleep(poll)


def wait_for_token(artifact_dir=None, timeout=30, poll=0.25):
    return wait_for(token_of, artifact_dir, timeout, poll)


def wait_for_body(substring, artifact_dir=None, timeout=30, poll=0.25):
    return wait_for(lambda body: body if substring in body else None,
                    artifact_dir, timeout, poll)
