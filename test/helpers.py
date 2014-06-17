import os
import ConfigParser
from redirect.util import create_token
from redirect.session_alchemy import mysql_spec
from redirect.models import User, Domain, Service

from redirect.util import hash

def mysql_spec_test():
    config = ConfigParser.ConfigParser()
    config_path = os.path.join(os.path.dirname(__file__), 'test_config.cfg')
    config.read(config_path)
    mysql_host = config.get('mysql', 'host')
    mysql_database = config.get('mysql', 'database')
    mysql_user = config.get('mysql', 'user')
    mysql_password = config.get('mysql', 'password')
    return mysql_spec(mysql_host, mysql_user, mysql_password, mysql_database)


def email():
    return unicode(create_token() + '@mail.com')

def domain():
    return create_token()

def token():
    return create_token()

def generate_user():
    uemail = email()
    activate_token = create_token()
    user = User(uemail, hash('pass1234'), False, activate_token)
    return user

def generate_domain():
    domain = create_token()
    update_token = create_token()
    domain = Domain(domain, None, update_token)
    return domain
