from flask import Flask, request, redirect, jsonify
from flask_cors import cross_origin
import db_helper
import services
from servicesexceptions import ServiceException, ParametersException
from dns import Dns
from mock import MagicMock
import sys, traceback
import json
import convertible
import config
import mail

import logging

the_config = config.read_redirect_configs()

app = Flask(__name__)

@app.route('/user/create', methods=["POST"])
@cross_origin()
def user_create():
    user = manager().create_new_user(request.form)
    user_data = convertible.to_dict(user)
    return jsonify(success=True, message='User was created', data=user_data), 200


@app.route('/user/activate', methods=["GET"])
@cross_origin()
def user_activate():
    manager().activate(request.args)
    return jsonify(success=True, message='User was activated'), 200


@app.route('/user/get', methods=["GET"])
@cross_origin()
def user_get():
    user = manager().authenticate(request.args)
    user_data = convertible.to_dict(user)
    return jsonify(success=True, message='User provided', data=user_data), 200


@app.route('/domain/acquire', methods=["POST"])
@cross_origin()
def domain_acquire():
    domain = manager().domain_acquire(request.form)
    return jsonify(success=True, user_domain=domain.user_domain, update_token=domain.update_token), 200


@app.route('/domain/drop_device', methods=["POST"])
@cross_origin()
def drop_device():
    domain = manager().drop_device(request.form)
    domain_data = convertible.to_dict(domain)
    return jsonify(success=True, message='Device was dropped', data=domain_data), 200


@app.route('/domain/get', methods=["GET"])
@cross_origin()
def domain_get():
    domain = manager().get_domain(request.args)
    domain_data = convertible.to_dict(domain)
    return jsonify(success=True, message='Domain retrieved', data=domain_data), 200


@app.route('/domain/update', methods=["POST"])
@cross_origin()
def domain_update():
    request_data = json.loads(request.data)
    domain = manager().domain_update(request_data, request.remote_addr)
    domain_data = convertible.to_dict(domain)
    return jsonify(success=True, message='Domain was updated', data=domain_data), 200


@app.route('/domain/delete', methods=["POST"])
@cross_origin()
def domain_delete():
    request_data = json.loads(request.data)
    manager().domain_delete(request_data)
    return jsonify(success=True, message='Domain was deleted'), 200


@app.route('/user/delete', methods=["POST"])
@cross_origin()
def user_delete():
    manager().delete_user(request.form)
    return jsonify(success=True, message='User deleted'), 200


@app.route('/user/reset_password', methods=["POST"])
@cross_origin()
def user_reset_password():
    manager().user_reset_password(request.form)
    return jsonify(success=True, message='Reset password requested'), 200


@app.route('/user/set_password', methods=["POST"])
@cross_origin()
def user_set_password():
    manager().user_set_password(request.form)
    return jsonify(success=True, message='Password was set successfully'), 200


@app.errorhandler(Exception)
@cross_origin()
def handle_exception(error):
    response = None
    status_code = 500
    if isinstance(error, ParametersException):
        parameters_messages = [{'parameter': k, 'messages': v} for k, v in error.parameters_errors.items()]
        response = jsonify(success=False, message=error.message, parameters_messages=parameters_messages)
        status_code = error.status_code
    elif isinstance(error, ServiceException):
        response = jsonify(success=False, message=error.message)
        status_code = error.status_code
    else:
        print '-'*60
        traceback.print_exc(file=sys.stdout)
        print '-'*60
        response = jsonify(success=False, message=error.message)
        status_code = 500
    logging.error(traceback.format_exc())
    logging.error(response.data)
    return response, status_code



def manager():
    the_config = config.read_redirect_configs()
    email_from = the_config.get('mail', 'from')
    activate_url_template = the_config.get('mail', 'activate_url_template')
    password_url_template = the_config.get('mail', 'password_url_template')

    redirect_domain = the_config.get('redirect', 'domain')
    redirect_activate_by_email = the_config.getboolean('redirect', 'activate_by_email')
    mock_dns = the_config.getboolean('redirect', 'mock_dns')

    if mock_dns:
        dns = MagicMock()
    else:
        dns = Dns(
            the_config.get('aws', 'access_key_id'),
            the_config.get('aws', 'secret_access_key'),
            the_config.get('aws', 'hosted_zone_id'))

    create_storage = db_helper.get_storage_creator(the_config)
    smtp = mail.get_smtp(the_config)

    the_mail = mail.Mail(smtp, email_from, activate_url_template, password_url_template)
    users_manager = services.Users(create_storage, redirect_activate_by_email, the_mail, dns, redirect_domain)
    return users_manager

if __name__ == '__main__':
    app.run(debug=True, port=5000)