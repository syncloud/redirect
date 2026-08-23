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


def get_token(body):
    return re.search(r'https://.*token=(.*)\r', body.replace('=\r\n', '')).group(1)
