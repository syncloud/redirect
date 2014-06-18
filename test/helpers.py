import os
import ConfigParser
from redirect.util import create_token, hash
from redirect.models import User, Domain, Service
from redirect.storage import get_session_maker, SessionContextFactory, mysql_spec_config


def get_storage_creator():
    config = get_test_config()
    spec = mysql_spec_config(config)
    maker = get_session_maker(spec)
    create_storage = SessionContextFactory(maker)
    return create_storage


def get_test_config():
    config = ConfigParser.ConfigParser()
    config_path = os.path.join(os.path.dirname(__file__), 'test_config.cfg')
    config.read(config_path)
    return config


def generate_user():
    email = unicode(create_token() + '@mail.com')
    activate_token = create_token()
    user = User(email, hash('pass1234'), False, activate_token)
    return user


def generate_domain():
    domain = create_token()
    update_token = create_token()
    domain = Domain(domain, None, update_token)
    return domain
