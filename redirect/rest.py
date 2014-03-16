import ConfigParser
import os
from flask import Flask, request, redirect
from dns import Dns
from db import Db
from mail import Mail
from accountmanager import AccountManager
from validation import Validator


config = ConfigParser.ConfigParser()
config.read(os.path.dirname(__file__) + '/config.cfg')

app = Flask(__name__)


@app.route('/')
def index():
    return redirect(manager().redirect_url(request, config.get('redirect', 'default_url')))


@app.route('/create')
def create():
    return manager().request_account(request)


@app.route('/activate')
def activate():
    return manager().activate(request)


@app.route('/update')
def update():
    return manager().update(request)


@app.route('/delete')
def delete():
    return manager().delete(request)


def manager():
    return AccountManager(
        Validator(),
        Db(
            config.get('mysql', 'host'),
            config.get('mysql', 'user'),
            config.get('mysql', 'passwd'),
            config.get('mysql', 'db')),
        Dns(
            config.get('aws', 'access_key_id'),
            config.get('aws', 'secret_access_key'),
            config.get('aws', 'hosted_zone_id')),
        config.get('redirect', 'domain'),
        config.getboolean('redirect', 'token_by_mail'),
        Mail(
            config.get('redirect', 'domain'),
            config.get('mail', 'from'))
    )

if __name__ == '__main__':
    app.run(debug=True)
