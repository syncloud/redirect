import db_helper
import services
from dns import Dns
from mock import MagicMock
import mail
import config
import statsd

class Ioc():
    def __init__(self, redirect_config = config.read_redirect_configs('/var/www/redirect')):
        self.statsd_client = statsd.StatsClient(redirect_config.get('stats', 'server'), 8125, prefix=the_config.get('stats', 'prefix'))
    
        from_email = redirect_config.get('mail', 'from')
        device_error_email = redirect_config.get('mail', 'device_error')
        activate_url_template = redirect_config.get('mail', 'activate_url_template')
        password_url_template = redirect_config.get('mail', 'password_url_template')

        redirect_domain = redirect_config.get('redirect', 'domain')
        redirect_activate_by_email = redirect_config.getboolean('redirect', 'activate_by_email')
        mock_dns = redirect_config.getboolean('redirect', 'mock_dns')

        if mock_dns:
            dns = MagicMock()
        else:
            statsd_cli = statsd_client(the_config)
            dns = Dns(
                redirect_config.get('aws', 'access_key_id'),
                redirect_config.get('aws', 'secret_access_key'),
                redirect_config.get('aws', 'hosted_zone_id'),
                statsd_cli)

        create_storage = db_helper.get_storage_creator(redirect_config)
        smtp = mail.get_smtp(the_config)

        the_mail = mail.Mail(smtp, from_email, activate_url_template, password_url_template, device_error_email)
        self.users_manager = services.Users(create_storage, redirect_activate_by_email, the_mail, dns, redirect_domain)
