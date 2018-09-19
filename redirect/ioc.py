import db_helper
import services
from dns import Dns
from mock import MagicMock
import mail
import config
import statsd

def statsd_client(the_config = config.read_redirect_configs()):
    return statsd.StatsClient(the_config.get('stats', 'server'), 8125, prefix=the_config.get('stats', 'prefix'))

def manager():
    the_config = config.read_redirect_configs()
    from_email = the_config.get('mail', 'from')
    device_error_email = the_config.get('mail', 'device_error')
    activate_url_template = the_config.get('mail', 'activate_url_template')
    password_url_template = the_config.get('mail', 'password_url_template')

    redirect_domain = the_config.get('redirect', 'domain')
    redirect_activate_by_email = the_config.getboolean('redirect', 'activate_by_email')
    mock_dns = the_config.getboolean('redirect', 'mock_dns')

    if mock_dns:
        dns = MagicMock()
    else:
        statsd_client = statsd_client(the_config)
        dns = Dns(
            the_config.get('aws', 'access_key_id'),
            the_config.get('aws', 'secret_access_key'),
            the_config.get('aws', 'hosted_zone_id'),
            statsd_client)

    create_storage = db_helper.get_storage_creator(the_config)
    smtp = mail.get_smtp(the_config)

    the_mail = mail.Mail(smtp, from_email, activate_url_template, password_url_template, device_error_email)
    users_manager = services.Users(create_storage, redirect_activate_by_email, the_mail, dns, redirect_domain)
    return users_manager