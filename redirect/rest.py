import json
import logging
import sys
import traceback

from flask import Flask, request, jsonify
from flask_cors import cross_origin
from syncloudlib.json import convertible

import ioc
from servicesexceptions import ServiceException, ParametersException

the_ioc = ioc.Ioc()
statsd_client = the_ioc.statsd_client
users_manager = the_ioc.users_manager
app = Flask(__name__)


@app.route('/user/get', methods=["GET"])
@cross_origin()
def user_get():
    statsd_client.incr('rest.user.get')
    user = users_manager.authenticate(request.args)
    user_data = convertible.to_dict(user)
    return jsonify(success=True, message='User provided', data=user_data), 200


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


@app.route('/user/log', methods=["POST"])
@cross_origin()
def user_log():
    statsd_client.incr('rest.user.log')
    users_manager.user_log(request.form)
    return jsonify(success=True, message='Error report sent successfully'), 200


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
