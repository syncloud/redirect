from flask import Flask, request, jsonify
from flask_cors import cross_origin
from servicesexceptions import ServiceException, ParametersException
import sys
import traceback
import json
from syncloudlib.json import convertible
import config
import logging
import ioc
from socket import gethostname

the_ioc = ioc.Ioc()
statsd_client = the_ioc.statsd_client
users_manager = the_ioc.users_manager
app = Flask(__name__)


@app.route('/user/activate', methods=["GET"])
@cross_origin()
def user_activate():
    statsd_client.incr('rest.user.activate')
    users_manager.activate(request.args)
    return jsonify(success=True, message='User was activated'), 200


@app.route('/user/get', methods=["GET"])
@cross_origin()
def user_get():
    statsd_client.incr('rest.user.get')
    user = users_manager.authenticate(request.args)
    user_data = convertible.to_dict(user)
    return jsonify(success=True, message='User provided', data=user_data), 200


@app.route('/user/create', methods=["POST"])
@cross_origin()
def user_create():
    statsd_client.incr('rest.user.create')
    user = users_manager.create_new_user(request.form)
    user_data = convertible.to_dict(user)
    return jsonify(success=True, message='User was created', data=user_data), 200


@app.route('/domain/acquire', methods=["POST"])
@cross_origin()
def domain_acquire():
    statsd_client.incr('rest.domain.acquire')
    domain = users_manager.domain_acquire(request.form)
    return jsonify(success=True, user_domain=domain.user_domain, update_token=domain.update_token), 200


@app.route('/domain/drop_device', methods=["POST"])
@cross_origin()
def drop_device():
    statsd_client.incr('rest.device.drop')
    domain = users_manager.drop_device(request.form)
    domain_data = convertible.to_dict(domain)
    return jsonify(success=True, message='Device was dropped', data=domain_data), 200


@app.route('/domain/delete', methods=["POST"])
@cross_origin()
def domain_delete():
    statsd_client.incr('rest.domain.delete')
    request_data = json.loads(request.data)
    users_manager.domain_delete(request_data)
    return jsonify(success=True, message='Domain was deleted'), 200


@app.route('/user/delete', methods=["POST"])
@cross_origin()
def user_delete():
    statsd_client.incr('rest.user.update')
    users_manager.delete_user(request.form)
    return jsonify(success=True, message='User deleted'), 200


@app.route('/user/log', methods=["POST"])
@cross_origin()
def user_log():
    statsd_client.incr('rest.user.log')
    users_manager.user_log(request.form)
    return jsonify(success=True, message='Error report sent successfully'), 200


@app.route('/probe/port', methods=["GET"])
@cross_origin()
def probe_port_v1():
    statsd_client.incr('rest.probe.port_v1')
    result, status_code = users_manager.port_probe(request.args, request.remote_addr)
    return result['message'], status_code


@app.route('/probe/port_v2', methods=["GET"])
@cross_origin()
def probe_port_v2():
    statsd_client.incr('rest.probe.port_v2')
    result, status_code = users_manager.port_probe(request.args, request.remote_addr)
    return json.dumps(result), status_code


@app.errorhandler(Exception)
@cross_origin()
def handle_exception(error):
    response = None
    status_code = 500
    logging.error('request.remote_addr: {0}'.format(request.remote_addr))
    logging.error('request.body: {0}'.format(request.data))
    if isinstance(error, ParametersException):
        statsd_client.incr('rest.exception.param')
        parameters_messages = [{'parameter': k, 'messages': v} for k, v in error.parameters_errors.items()]
        response = jsonify(success=False, message=error.message, parameters_messages=parameters_messages)
        status_code = error.status_code
    elif isinstance(error, ServiceException):
        statsd_client.incr('rest.exception.service')
        response = jsonify(success=False, message=error.message)
        status_code = error.status_code
    else:
        statsd_client.incr('rest.exception.unknown')
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
