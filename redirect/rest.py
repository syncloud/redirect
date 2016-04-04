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
import graphitesend

the_config = config.read_redirect_configs()
graphitesend.init(graphite_server=the_config.get('stats', 'server'))

app = Flask(__name__)


@app.route('/user/create', methods=["POST"])
@cross_origin()
def user_create():
    graphitesend.send('rest.user.create', 1)
    user = ioc.manager().create_new_user(request.form)
    user_data = convertible.to_dict(user)
    return jsonify(success=True, message='User was created', data=user_data), 200


@app.route('/user/activate', methods=["GET"])
@cross_origin()
def user_activate():
    graphitesend.send('rest.user.activate', 1)
    ioc.manager().activate(request.args)
    return jsonify(success=True, message='User was activated'), 200


@app.route('/user/get', methods=["GET"])
@cross_origin()
def user_get():
    graphitesend.send('rest.user.get', 1)
    user = ioc.manager().authenticate(request.args)
    user_data = convertible.to_dict(user)
    return jsonify(success=True, message='User provided', data=user_data), 200


@app.route('/domain/acquire', methods=["POST"])
@cross_origin()
def domain_acquire():
    graphitesend.send('rest.domain.acquire', 1)
    domain = ioc.manager().domain_acquire(request.form)
    return jsonify(success=True, user_domain=domain.user_domain, update_token=domain.update_token), 200


@app.route('/domain/drop_device', methods=["POST"])
@cross_origin()
def drop_device():
    graphitesend.send('rest.device.drop', 1)
    domain = ioc.manager().drop_device(request.form)
    domain_data = convertible.to_dict(domain)
    return jsonify(success=True, message='Device was dropped', data=domain_data), 200


@app.route('/domain/get', methods=["GET"])
@cross_origin()
def domain_get():
    graphitesend.send('rest.domain.get', 1)
    domain = ioc.manager().get_domain(request.args)
    domain_data = convertible.to_dict(domain)
    return jsonify(success=True, message='Domain retrieved', data=domain_data), 200


@app.route('/domain/update', methods=["POST"])
@cross_origin()
def domain_update():
    graphitesend.send('rest.domain.update', 1)
    request_data = json.loads(request.data)
    domain = ioc.manager().domain_update(request_data, request.remote_addr)
    domain_data = convertible.to_dict(domain)
    return jsonify(success=True, message='Domain was updated', data=domain_data), 200


@app.route('/domain/delete', methods=["POST"])
@cross_origin()
def domain_delete():
    graphitesend.send('rest.domain.delete', 1)
    request_data = json.loads(request.data)
    ioc.manager().domain_delete(request_data)
    return jsonify(success=True, message='Domain was deleted'), 200


@app.route('/user/delete', methods=["POST"])
@cross_origin()
def user_delete():
    graphitesend.send('rest.user.update', 1)
    ioc.manager().delete_user(request.form)
    return jsonify(success=True, message='User deleted'), 200


@app.route('/user/reset_password', methods=["POST"])
@cross_origin()
def user_reset_password():
    graphitesend.send('rest.user.reset_password', 1)
    ioc.manager().user_reset_password(request.form)
    return jsonify(success=True, message='Reset password requested'), 200


@app.route('/user/log', methods=["POST"])
@cross_origin()
def user_log():
    graphitesend.send('rest.user.log', 1)
    ioc.manager().user_log(request.form)
    return jsonify(success=True, message='Error report sent successfully'), 200


@app.route('/user/set_password', methods=["POST"])
@cross_origin()
def user_set_password():
    graphitesend.send('rest.user.set_password', 1)
    ioc.manager().user_set_password(request.form)
    return jsonify(success=True, message='Password was set successfully'), 200


@app.route('/probe/port', methods=["GET"])
@cross_origin()
def probe_port():
    graphitesend.send('rest.probe.port', 1)
    return ioc.manager().port_probe(request.args, request.remote_addr)


@app.errorhandler(Exception)
@cross_origin()
def handle_exception(error):
    response = None
    status_code = 500
    logging.error('request.remote_addr: {0}'.format(request.remote_addr))
    logging.error('request.body: {0}'.format(request.data))
    if isinstance(error, ParametersException):
        graphitesend.send('rest.exception.param', 1)
        parameters_messages = [{'parameter': k, 'messages': v} for k, v in error.parameters_errors.items()]
        response = jsonify(success=False, message=error.message, parameters_messages=parameters_messages)
        status_code = error.status_code
    elif isinstance(error, ServiceException):
        graphitesend.send('rest.exception.service', 1)
        response = jsonify(success=False, message=error.message)
        status_code = error.status_code
    else:
        graphitesend.send('rest.exception.unknown', 1)
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
