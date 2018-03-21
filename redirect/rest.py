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
import statsd
from socket import gethostname

the_config = config.read_redirect_configs()
statsd_client = statsd.StatsClient(the_config.get('stats', 'server'), 8125, prefix=gethostname())

app = Flask(__name__)


@app.route('/user/create', methods=["POST"])
@cross_origin()
def user_create():
    statsd_client.incr('rest.user.create')
    user = ioc.manager().create_new_user(request.form)
    user_data = convertible.to_dict(user)
    return jsonify(success=True, message='User was created', data=user_data), 200


@app.route('/user/activate', methods=["GET"])
@cross_origin()
def user_activate():
    statsd_client.incr('rest.user.activate')
    ioc.manager().activate(request.args)
    return jsonify(success=True, message='User was activated'), 200


@app.route('/user/get', methods=["GET"])
@cross_origin()
def user_get():
    statsd_client.incr('rest.user.get')
    user = ioc.manager().authenticate(request.args)
    user_data = convertible.to_dict(user)
    return jsonify(success=True, message='User provided', data=user_data), 200


@app.route('/domain/acquire', methods=["POST"])
@cross_origin()
def domain_acquire():
    statsd_client.incr('rest.domain.acquire')
    domain = ioc.manager().domain_acquire(request.form)
    return jsonify(success=True, user_domain=domain.user_domain, update_token=domain.update_token), 200


@app.route('/domain/drop_device', methods=["POST"])
@cross_origin()
def drop_device():
    statsd_client.incr('rest.device.drop')
    domain = ioc.manager().drop_device(request.form)
    domain_data = convertible.to_dict(domain)
    return jsonify(success=True, message='Device was dropped', data=domain_data), 200


@app.route('/domain/get', methods=["GET"])
@cross_origin()
def domain_get():
    statsd_client.incr('rest.domain.get')
    domain = ioc.manager().get_domain(request.args)
    domain_data = convertible.to_dict(domain)
    return jsonify(success=True, message='Domain retrieved', data=domain_data), 200


@app.route('/domain/update', methods=["POST"])
@cross_origin()
def domain_update():
    statsd_client.incr('rest.domain.update')
    request_data = json.loads(request.data)
    domain = ioc.manager().domain_update(request_data, request.remote_addr)
    domain_data = convertible.to_dict(domain)
    return jsonify(success=True, message='Domain was updated', data=domain_data), 200


@app.route('/domain/delete', methods=["POST"])
@cross_origin()
def domain_delete():
    statsd_client.incr('rest.domain.delete')
    request_data = json.loads(request.data)
    ioc.manager().domain_delete(request_data)
    return jsonify(success=True, message='Domain was deleted'), 200


@app.route('/user/delete', methods=["POST"])
@cross_origin()
def user_delete():
    statsd_client.incr('rest.user.update')
    ioc.manager().delete_user(request.form)
    return jsonify(success=True, message='User deleted'), 200


@app.route('/user/reset_password', methods=["POST"])
@cross_origin()
def user_reset_password():
    statsd_client.incr('rest.user.reset_password')
    ioc.manager().user_reset_password(request.form)
    return jsonify(success=True, message='Reset password requested'), 200


@app.route('/user/log', methods=["POST"])
@cross_origin()
def user_log():
    statsd_client.incr('rest.user.log')
    ioc.manager().user_log(request.form)
    return jsonify(success=True, message='Error report sent successfully'), 200


@app.route('/user/set_password', methods=["POST"])
@cross_origin()
def user_set_password():
    statsd_client.incr('rest.user.set_password')
    ioc.manager().user_set_password(request.form)
    return jsonify(success=True, message='Password was set successfully'), 200


@app.route('/probe/port', methods=["GET"])
@cross_origin()
def probe_port_v1():
    statsd_client.incr('rest.probe.port_v1')
    result, status_code = ioc.manager().port_probe(request.args, request.remote_addr)
    return result['message'], status_code


@app.route('/probe/port_v2', methods=["GET"])
@cross_origin()
def probe_port_v2():
    statsd_client.incr('rest.probe.port_v2')
    return ioc.manager().port_probe(request.args, request.remote_addr)


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
