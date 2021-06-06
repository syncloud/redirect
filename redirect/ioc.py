import statsd

import config
import db_helper
import mail
import services


class Ioc:
    def __init__(self, redirect_config=config.read_redirect_configs('/var/www/redirect')):
        self.redirect_config = redirect_config
        self.statsd_client = statsd.StatsClient(
            redirect_config.get('stats', 'server'), 8125,
            prefix=redirect_config.get('stats', 'prefix'))
    
        from_email = redirect_config.get('mail', 'from')
        device_error_email = redirect_config.get('mail', 'device_error')
        activate_url_template = redirect_config.get('mail', 'activate_url_template')
        password_url_template = redirect_config.get('mail', 'password_url_template')

        redirect_activate_by_email = redirect_config.getboolean('redirect', 'activate_by_email')

        create_storage = db_helper.get_storage_creator(redirect_config)
        smtp = mail.get_smtp(redirect_config)

        the_mail = mail.Mail(smtp, from_email, activate_url_template, password_url_template, device_error_email)
        self.users_manager = services.Users(create_storage, redirect_activate_by_email, the_mail)
