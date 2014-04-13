import ConfigParser
import os
from flask import Flask, request, redirect, jsonify, send_from_directory
from flask.ext.login import LoginManager, login_user, logout_user, current_user, login_required
import services
import storage
from servicesexceptions import ServiceException
from mail import Mail
from dns import Dns
from mock import MagicMock

from flaskutil import crossdomain

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
@crossdomain(origin='*')
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
@crossdomain(origin='*')
@login_required
def user():
    user = current_user.user
    return jsonify(email=user.email, user_domain=user.user_domain, ip=user.ip, port=user.port)

@app.route('/user/create', methods=["POST"])
@crossdomain(origin='*')
def user_create():
    manager().create_new_user(request.form)
    return 'User was created', 200

@app.route('/user/activate', methods=["GET"])
def user_activate():
    manager().activate(request.args)
    return 'User was activated', 200

@app.route('/user/get', methods=["GET"])
def user_get():
    user = manager().authenticate(request.args)
    return jsonify(user_domain=user.user_domain, update_token=user.update_token, ip=user.ip, port=user.port, email=user.email, active=user.active)

@app.route('/domain/update', methods=["POST"])
def update_ip_port():
    manager().update_ip_port(request.form)
    return 'Domain was updated', 200

@app.errorhandler(Exception)
@crossdomain(origin='*')
def handle_exception(error):
    if isinstance(error, ServiceException):
        return error.message, error.status_code
    else:
        return error.message, 500

@app.errorhandler(401)
@crossdomain(origin='*')
def handle_401(e):
        return e, 401

def manager():
    mysql_host = config.get('mysql', 'host')
    mysql_user = config.get('mysql', 'user')
    mysql_password = config.get('mysql', 'passwd')
    mysql_db = config.get('mysql', 'db')

    mail_host = config.get('smtp', 'host')
    mail_port = config.get('smtp', 'port')

    mail_from = config.get('mail', 'from')

    redirect_domain = config.get('redirect', 'domain')
    redirect_activate_by_email = config.get('redirect', 'activate_by_email').lower() != 'false'
    activate_url_template = config.get('redirect', 'activate_url_template')
    mock_dns = bool(config.get('redirect', 'mock_dns'))

    if mock_dns:
        dns = MagicMock()
    else:
        dns = Dns(
            config.get('aws', 'access_key_id'),
            config.get('aws', 'secret_access_key'),
            config.get('aws', 'hosted_zone_id'))


    user_storage = storage.UserStorage(mysql_host, mysql_user, mysql_password, mysql_db)
    mail = Mail(mail_host, mail_port, mail_from)
    users_manager = services.Users(user_storage, redirect_activate_by_email, mail, activate_url_template, dns, redirect_domain)
    return users_manager

if __name__ == '__main__':
    app.run(debug=True)
