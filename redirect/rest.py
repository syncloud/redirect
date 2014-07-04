import ConfigParser
import os
from flask import Flask, request, redirect, jsonify
from flask_cors import cross_origin
import db_helper
import services
from servicesexceptions import ServiceException
from mail import Mail
from dns import Dns
from mock import MagicMock

import models

import json
import jsonpickle

config = ConfigParser.ConfigParser()
config.read(os.path.join(os.path.dirname(__file__), 'config.cfg'))

app = Flask(__name__)


@app.route('/')
def index():
    return redirect(manager().redirect_url(request.url))


@app.route('/user/create', methods=["POST"])
@cross_origin()
def user_create():
    manager().create_new_user(request.form)
    return jsonify(message='User was created'), 200


@app.route('/user/activate', methods=["GET"])
@cross_origin()
def user_activate():
    manager().activate(request.args)
    return jsonify(message='User was activated'), 200


@app.route('/user/get', methods=["GET"])
@cross_origin()
def user_get():
    user = manager().authenticate(request.args)
    domain = user.domains[0]
    port = None
    if len(domain.services) == 1:
        service = domain.services[0]
        port = service.port
    return jsonify(user_domain=domain.user_domain, update_token=domain.update_token, ip=domain.ip,
                   port=port, email=user.email, active=user.active), 200


@app.route('/domain/acquire', methods=["POST"])
@cross_origin()
def domain_acquire():
    domain = manager().domain_acquire(request.form)
    return jsonify(user_domain=domain.user_domain, update_token=domain.update_token), 200


@app.route('/domain/get', methods=["GET"])
@cross_origin()
def domain_get():
    domain = manager().get_domain(request.args)
    domain_data = models.to_dict(domain)
    return jsonify(message='Domain retrieved', data=domain_data), 200


@app.route('/domain/update', methods=["POST"])
@cross_origin()
def domain_update2():
    request_data = json.loads(request.data)
    domain = manager().domain_update(request_data)
    domain_data = models.to_dict(domain)
    return jsonify(message='Domain was updated', data=domain_data), 200


@app.route('/user/delete', methods=["POST"])
@cross_origin()
def user_delete():
    manager().delete_user(request.form)
    return jsonify(message='User deleted'), 200


@app.errorhandler(Exception)
@cross_origin()
def handle_exception(error):
    if isinstance(error, ServiceException):
        return jsonify(message=error.message), error.status_code
    else:
        return jsonify(message=error.message), 500


def manager():
    mail_host = config.get('smtp', 'host')
    mail_port = config.get('smtp', 'port')

    mail_from = config.get('mail', 'from')

    redirect_domain = config.get('redirect', 'domain')
    redirect_activate_by_email = config.getboolean('redirect', 'activate_by_email')
    activate_url_template = config.get('redirect', 'activate_url_template')
    mock_dns = config.getboolean('redirect', 'mock_dns')

    if mock_dns:
        dns = MagicMock()
    else:
        dns = Dns(
            config.get('aws', 'access_key_id'),
            config.get('aws', 'secret_access_key'),
            config.get('aws', 'hosted_zone_id'))

    create_storage = db_helper.get_storage_creator(config)

    mail = Mail(mail_host, mail_port, mail_from)
    users_manager = services.Users(create_storage, redirect_activate_by_email,
                                   mail, activate_url_template, dns, redirect_domain)
    return users_manager

if __name__ == '__main__':
    app.run(debug=True, port=5000)