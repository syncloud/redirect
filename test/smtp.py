import json
from urlparse import urlparse

import requests


def emails():
    response = requests.get('http://mail:8025/api/v1/messages')
    assert response.status_code == 200, response.text
    return [message['Content'] for message in json.loads(response.text)]


def email_bodies():
    return [message['Body'] for message in emails()]


def clear():
    response = requests.delete('http://mail:8025/api/v1/messages')
    assert response.status_code == 200, response.text


def get_token(body):
    link_index = body.find('http://')
    link = body[link_index:].split(' ')[0].strip()
    parts = urlparse(link)
    token = parts.query.replace('token=', '')
    return token
