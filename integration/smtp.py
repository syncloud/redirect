import json

import requests


def emails():
    response = requests.get('http://mail:8025/api/v1/messages')
    assert response.status_code == 200, response.text
    return [message['body'] for message in json.loads(response.text)]


def clear():
    response = requests.delete('http://mail:8025/api/v1/messages')
    assert response.status_code == 200, response.text

