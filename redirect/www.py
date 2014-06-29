import ConfigParser
import os
from flask import Flask, request, redirect, jsonify, send_from_directory
from flask.ext.login import LoginManager, login_user, logout_user, current_user, login_required
from redirect.db_helper import get_storage_creator
import services
from servicesexceptions import ServiceException
from mail import Mail
from dns import Dns
from mock import MagicMock

config = ConfigParser.ConfigParser()
config.read(os.path.join(os.path.dirname(__file__), 'config.cfg'))

app = Flask(__name__)
app.config['SECRET_KEY'] = config.get('redirect', 'auth_secret_key')
login_manager = LoginManager()
login_manager.init_app(app)

# This is to host static html - should be done on proper web server in prod

host_static_files = config.has_option('redirect', 'static_files_path')
if host_static_files:
    static_files_path = config.get('redirect', 'static_files_path')

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

@app.route('/')
def index():
    return redirect(manager().redirect_url(request.url))

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
    return jsonify(email=user.email, user_domain=user.user_domain, ip=user.ip, port=user.port, update_token=user.update_token)

@app.errorhandler(Exception)
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
    redirect_activate_by_email = bool(config.get('redirect', 'activate_by_email'))
    activate_url_template = config.get('redirect', 'activate_url_template')
    mock_dns = bool(config.get('redirect', 'mock_dns'))

    if mock_dns:
        dns = MagicMock()
    else:
        dns = Dns(
            config.get('aws', 'access_key_id'),
            config.get('aws', 'secret_access_key'),
            config.get('aws', 'hosted_zone_id'))

    create_storage = get_storage_creator(config)

    mail = Mail(mail_host, mail_port, mail_from)
    users_manager = services.Users(create_storage, redirect_activate_by_email, mail, activate_url_template, dns, redirect_domain)
    return users_manager

if __name__ == '__main__':
    app.run(debug=True, port=5001)
