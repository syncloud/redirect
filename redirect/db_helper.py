from redirect.storage import mysql_spec_config, get_session_maker, SessionContextFactory


def get_storage_creator(config):
    spec = mysql_spec_config(config)
    maker = get_session_maker(spec)
    create_storage = SessionContextFactory(maker)
    return create_storage
