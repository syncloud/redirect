import traceback

from flask import Flask, request, jsonify
from flask_login import LoginManager, login_user, logout_user, current_user, login_required
from syncloudlib.json import convertible

import ioc
from backend_proxy import backend_request
from servicesexceptions import ServiceException, ParametersException

the_ioc = ioc.Ioc()
statsd_client = the_ioc.statsd_client
users_manager = the_ioc.users_manager

app = Flask(__name__)
app.config['SECRET_KEY'] = the_ioc.redirect_config.get('redirect', 'auth_secret_key')
login_manager = LoginManager()
login_manager.init_app(app)


class UserFlask:
    def __init__(self, user):
        self.user = user

    def is_authenticated(self):
        return True

    def is_active(self):
        return True

    def is_anonymous(self):
        return False

    def get_id(self):
        return unicode(self.user)


@login_manager.user_loader
def load_user(email):
    user = users_manager.get_user(email)
    if not user:
        return None
    return UserFlask(user)


@app.route("/login", methods=["POST"])
def login():
    statsd_client.incr('www.user.login')
    user = users_manager.authenticate(request.form)
    user_flask = UserFlask(user.email)
    login_user(user_flask, remember=False)
    return 'User logged in', 200


@app.route("/logout", methods=["POST"])
@login_required
def logout():
    statsd_client.incr('www.user.logout')
    logout_user()
    return 'User logged out', 200


@app.route('/user/set_password', methods=["POST"])
def user_set_password():
    statsd_client.incr('rest.user.set_password')
    users_manager.user_set_password(request.form)
    return jsonify(success=True, message='Password was set successfully'), 200


@app.route('/user/reset_password', methods=["POST"])
@app.route('/user/activate', methods=["POST"])
@app.route('/user/create', methods=["POST"])
def backend_proxy_public():
    response = backend_request(request.method, '/web' + request.full_path, request.json, headers={})
    return response.text, response.status_code


@app.route("/notification/subscribe", methods=["POST"])
@app.route("/notification/unsubscribe", methods=["POST"])
@app.route("/user", methods=["GET"])
@app.route("/domains", methods=["GET"])
@app.route("/premium/request", methods=["POST"])
@app.route("/domain", methods=["DELETE"])
@login_required
def backend_proxy_private():
    response = backend_request(request.method, '/web' + request.full_path, request.json,
                               headers={'RedirectUserEmail': current_user.user.email})
    return response.text, response.status_code


@app.route("/user", methods=["DELETE"])
@login_required
def backend_proxy_user_delete():
    response = backend_request(request.method, '/web' + request.full_path, request.json,
                               headers={'RedirectUserEmail': current_user.user.email})
    if response.status_code != 200:
        return response.text, response.status_code
    else:
        logout_user()
    return 'OK'


@app.errorhandler(Exception)
def handle_exception(error):
    if isinstance(error, ParametersException):
        statsd_client.incr('www.exception.param')
        parameters_messages = [{'parameter': k, 'messages': v} for k, v in error.parameters_errors.items()]
        return jsonify(message=error.message, parameters_messages=parameters_messages), error.status_code
    if isinstance(error, ServiceException):
        statsd_client.incr('www.exception.service')
        return jsonify(message=error.message), error.status_code
    else:
        statsd_client.incr('www.exception.unknown')
        tb = traceback.format_exc()
        return jsonify(message=tb), 500


if __name__ == '__main__':
    app.run(debug=True, port=5001)
