from flask import Flask, request, redirect, jsonify, send_from_directory
from flask.ext.login import LoginManager, login_user, logout_user, current_user, login_required
import db_helper
import services
from servicesexceptions import ServiceException, ParametersException
from dns import Dns
from mock import MagicMock
import traceback
import convertible
import config
import mail

the_config = config.read_redirect_configs()

app = Flask(__name__)
app.config['SECRET_KEY'] = the_config.get('redirect', 'auth_secret_key')
login_manager = LoginManager()
login_manager.init_app(app)

# This is to host static html - should be done on proper web server in prod

host_static_files = the_config.has_option('redirect', 'static_files_path')
if host_static_files:
    static_files_path = the_config.get('redirect', 'static_files_path')

    @app.route('/<path:filename>')
    def static_file(filename):
        return send_from_directory(static_files_path, filename)

# End of hosting static html


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
    user = (manager().get_user(email))
    if not user:
        return None
    return UserFlask(user)

@app.route("/login", methods=["POST"])
def login():
    user = manager().authenticate(request.form)
    user_flask = UserFlask(user)
    login_user(user_flask, remember=False)
    return 'User logged in', 200

@app.route("/logout", methods=["POST"])
@login_required
def logout():
    logout_user()
    return 'User logged out', 200

@app.route("/user", methods=["GET"])
@login_required
def user():
    user = current_user.user
    user_data = convertible.to_dict(user)
    return jsonify(user_data), 200

@app.route("/set_subscribed", methods=["POST"])
@login_required
def user_unsubscribe():
    user = current_user.user
    manager().user_set_subscribed(request.form, user.email)
    return 'Successfully set', 200

@app.route("/domain_delete", methods=["POST"])
@login_required
def domain_delete():
    user = current_user.user
    manager().user_domain_delete(request.form, user)
    return 'Domain deleted', 200

@app.errorhandler(Exception)
def handle_exception(error):
    if isinstance(error, ParametersException):
        parameters_messages = [{'parameter': k, 'messages': v} for k, v in error.parameters_errors.items()]
        return jsonify(message=error.message, parameters_messages=parameters_messages), error.status_code
    if isinstance(error, ServiceException):
        return jsonify(message=error.message), error.status_code
    else:
        tb = traceback.format_exc()
        return jsonify(message=tb), 500


# def manager():
#     the_config = config.read_redirect_configs()
#     redirect_domain = the_config.get('redirect', 'domain')
#     create_storage = db_helper.get_storage_creator(the_config)
#     users_manager = services.Users(create_storage, None, None, None, redirect_domain)
#     return users_manager


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
    app.run(debug=True, port=5001)
