import traceback

from flask import Flask, request, jsonify
from flask_login import LoginManager, login_user, logout_user, current_user, login_required
from syncloudlib.json import convertible

import ioc
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
        return self.user.active

    def is_anonymous(self):
        return False

    def get_id(self):
        return unicode(self.user.email)


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
    user_flask = UserFlask(user)
    login_user(user_flask, remember=False)
    return 'User logged in', 200


@app.route("/logout", methods=["POST"])
@login_required
def logout():
    statsd_client.incr('www.user.logout')
    logout_user()
    return 'User logged out', 200


@app.route("/user/get", methods=["GET"])
@login_required
def user():
    statsd_client.incr('www.user.get')
    user = current_user.user
    user_data = convertible.to_dict(user)
    return jsonify(user_data), 200


@app.route('/user/create', methods=["POST"])
def user_create():
    statsd_client.incr('www.user.create')
    user = users_manager.create_new_user(request.form)
    user_data = convertible.to_dict(user)
    return jsonify(success=True, message='User was created', data=user_data), 200


@app.route('/user/reset_password', methods=["POST"])
def user_reset_password():
    statsd_client.incr('www.user.reset_password')
    users_manager.user_reset_password(request.form)
    return jsonify(success=True, message='Reset password requested'), 200


@app.route('/user/set_password', methods=["POST"])
def user_set_password():
    statsd_client.incr('rest.user.set_password')
    users_manager.user_set_password(request.form)
    return jsonify(success=True, message='Password was set successfully'), 200


@app.route('/user/activate', methods=["POST"])
def user_activate():
    statsd_client.incr('rest.user.activate')
    users_manager.activate(request.form)
    return jsonify(success=True, message='User was activated'), 200


@app.route("/user_delete", methods=["POST"])
@login_required
def user_delete():
    statsd_client.incr('www.user.delete')
    user = current_user.user
    users_manager.do_delete_user(user.email)
    return 'User deleted', 200


@app.route("/set_subscribed", methods=["POST"])
@login_required
def user_unsubscribe():
    statsd_client.incr('www.user.unsubscribe')
    user = current_user.user
    users_manager.user_set_subscribed(request.form, user.email)
    return 'Successfully set', 200


@app.route("/domain_delete", methods=["POST"])
@login_required
def domain_delete():
    statsd_client.incr('www.domain.delete')
    user = current_user.user
    users_manager.user_domain_delete(request.form, user)
    return 'Domain deleted', 200


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
