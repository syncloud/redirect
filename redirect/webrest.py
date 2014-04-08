import ConfigParser
import os
from flask import Flask, request, redirect, jsonify
from flask.ext.login import LoginManager, login_user, logout_user, current_user, login_required
import services
import storage
from servicesexceptions import ServiceException
from mail import Mail

config = ConfigParser.ConfigParser()
config.read(os.path.join(os.path.dirname(__file__), 'config.cfg'))

app = Flask(__name__)
app.config['SECRET_KEY'] = config.get('redirect', 'auth_secret_key')
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
    user = (manager().get_user(email))
    if not user:
        return None
    return UserFlask(user)

@app.route('/')
def index():
    return redirect(manager().redirect_url(request, config.get('redirect', 'default_url')))

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

@app.route("/user", methods=["POST"])
@login_required
def user():
    user = current_user.user
    return jsonify(email=user.email, user_domain=user.user_domain, ip=user.ip, port=user.port)

@app.route('/user/create', methods=["POST"])
def create():
    manager().create_new_user(request.form)
    return 'User was created', 200

@app.route('/user/activate', methods=["GET"])
def activate():
    manager().activate(request.args)
    return 'User was activated', 200

@app.errorhandler(Exception)
def handle_exception(error):
    if error is ServiceException:
        return error.message, error.status_code
    else:
        return error.message, 500

def manager():
    mysql_host = config.get('mysql', 'host')
    mysql_user = config.get('mysql', 'user')
    mysql_password = config.get('mysql', 'passwd')
    mysql_db = config.get('mysql', 'db')

    mail_host = config.get('smtp', 'host')
    mail_port = config.get('smtp', 'port')

    mail_from = config.get('mail', 'from')

    redirect_domain = config.get('redirect', 'domain')
    activate_url_template = config.get('redirect', 'activate_url_template')

    user_storage = storage.UserStorage(mysql_host, mysql_user, mysql_password, mysql_db)
    mail = Mail(mail_host, mail_port, redirect_domain, mail_from)
    users_manager = services.Users(user_storage, mail, activate_url_template)
    return users_manager

if __name__ == '__main__':
    app.run(debug=True)
