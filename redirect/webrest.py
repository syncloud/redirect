import ConfigParser
import os
from flask import Flask, request, redirect
import services
import storage
from servicesexceptions import ServiceException
from mail import Mail

app = Flask(__name__)

config = ConfigParser.ConfigParser()
config.read(os.path.join(os.path.dirname(__file__), 'config.cfg'))


@app.route('/')
def index():
    return redirect(manager().redirect_url(request, config.get('redirect', 'default_url')))

@app.route('/user/create', methods=["POST"])
def create():
    user = manager().create_new_user(request.form)
    return 'User was created', 200

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
