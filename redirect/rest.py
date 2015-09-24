from flask import Flask, request, jsonify
from flask_cors import cross_origin
from servicesexceptions import ServiceException, ParametersException
import sys
import traceback
import json
import convertible
import config
import logging
import ioc

the_config = config.read_redirect_configs()

app = Flask(__name__)

@app.route('/user/create', methods=["POST"])
@cross_origin()
def user_create():
    user = ioc.manager().create_new_user(request.form)
    user_data = convertible.to_dict(user)
    return jsonify(success=True, message='User was created', data=user_data), 200


@app.route('/user/activate', methods=["GET"])
@cross_origin()
def user_activate():
    ioc.manager().activate(request.args)
    return jsonify(success=True, message='User was activated'), 200


@app.route('/user/get', methods=["GET"])
@cross_origin()
def user_get():
    user = ioc.manager().authenticate(request.args)
    user_data = convertible.to_dict(user)
    return jsonify(success=True, message='User provided', data=user_data), 200


@app.route('/domain/acquire', methods=["POST"])
@cross_origin()
def domain_acquire():
    domain = ioc.manager().domain_acquire(request.form)
    return jsonify(success=True, user_domain=domain.user_domain, update_token=domain.update_token), 200


@app.route('/domain/drop_device', methods=["POST"])
@cross_origin()
def drop_device():
    domain = ioc.manager().drop_device(request.form)
    domain_data = convertible.to_dict(domain)
    return jsonify(success=True, message='Device was dropped', data=domain_data), 200


@app.route('/domain/get', methods=["GET"])
@cross_origin()
def domain_get():
    domain = ioc.manager().get_domain(request.args)
    domain_data = convertible.to_dict(domain)
    return jsonify(success=True, message='Domain retrieved', data=domain_data), 200


@app.route('/domain/update', methods=["POST"])
@cross_origin()
def domain_update():
    request_data = json.loads(request.data)
    domain = ioc.manager().domain_update(request_data, request.remote_addr)
    domain_data = convertible.to_dict(domain)
    return jsonify(success=True, message='Domain was updated', data=domain_data), 200


@app.route('/domain/delete', methods=["POST"])
@cross_origin()
def domain_delete():
    request_data = json.loads(request.data)
    ioc.manager().domain_delete(request_data)
    return jsonify(success=True, message='Domain was deleted'), 200


@app.route('/user/delete', methods=["POST"])
@cross_origin()
def user_delete():
    ioc.manager().delete_user(request.form)
    return jsonify(success=True, message='User deleted'), 200


@app.route('/user/reset_password', methods=["POST"])
@cross_origin()
def user_reset_password():
    ioc.manager().user_reset_password(request.form)
    return jsonify(success=True, message='Reset password requested'), 200


@app.route('/user/log', methods=["POST"])
@cross_origin()
def user_log():
    ioc.manager().user_log(request.form)
    return jsonify(success=True, message='Error report sent successfully'), 200


@app.route('/user/set_password', methods=["POST"])
@cross_origin()
def user_set_password():
    ioc.manager().user_set_password(request.form)
    return jsonify(success=True, message='Password was set successfully'), 200


@app.route('/probe/port', methods=["GET"])
@cross_origin()
def probe_port():
    return ioc.manager().port_probe(request.args)

@app.errorhandler(Exception)
@cross_origin()
def handle_exception(error):
    response = None
    status_code = 500
    logging.error('request.remote_addr: {0}'.format(request.remote_addr))
    logging.error('request.body: {0}'.format(request.data))
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


if __name__ == '__main__':
    app.run(debug=True, port=5000)
