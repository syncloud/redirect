import ConfigParser
import os
from flask import Flask, request, redirect
import users
import storage
from restexceptions import RestException

app = Flask(__name__)

config = ConfigParser.ConfigParser()
config.read(os.path.join(os.path.dirname(__file__), 'config.cfg'))


@app.route('/')
def index():
    return redirect(manager().redirect_url(request, config.get('redirect', 'default_url')))

@app.route('/user/create', methods=["POST"])
def create():
    user = manager().create_new_user(request.form)
    return 'User was created', 200, {'Token': user.activate_token}

@app.errorhandler(RestException)
def handle_invalid_usage(error):
    return (error.message, error.status_code)

def manager():
    mysql_host = config.get('mysql', 'host')
    mysql_user = config.get('mysql', 'user')
    mysql_password = config.get('mysql', 'passwd')
    mysql_db = config.get('mysql', 'db')

    user_storage = storage.UserStorage(mysql_host, mysql_user, mysql_password, mysql_db)
    users_manager = users.Users(user_storage)
    return users_manager

if __name__ == '__main__':
    app.run(debug=True)
