import statsd

import config
import db_helper
import services


class Ioc:
    def __init__(self, redirect_config=config.read_redirect_configs('/var/www/redirect')):
        self.redirect_config = redirect_config
        self.statsd_client = statsd.StatsClient(
            redirect_config.get('stats', 'server'), 8125,
            prefix=redirect_config.get('stats', 'prefix'))
    
        create_storage = db_helper.get_storage_creator(redirect_config)
        self.users_manager = services.Users(create_storage)
