from flask import Flask, request, redirect, jsonify, send_from_directory
from flask.ext.login import LoginManager, login_user, logout_user, current_user, login_required
import db_helper
import services
from servicesexceptions import ServiceException
import traceback
import convertible
import config

config = config.read_redirect_configs()

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
    user_data = convertible.to_dict(user)
    return jsonify(user_data), 200

@app.errorhandler(Exception)
def handle_exception(error):
    if isinstance(error, ServiceException):
        return jsonify(message=error.message), error.status_code
    else:
        tb = traceback.format_exc()
        return jsonify(message=tb), 500


def manager():
    redirect_domain = config.get('redirect', 'domain')
    create_storage = db_helper.get_storage_creator(config)
    users_manager = services.UsersRead(create_storage, redirect_domain)
    return users_manager

if __name__ == '__main__':
    app.run(debug=True, port=5001)
