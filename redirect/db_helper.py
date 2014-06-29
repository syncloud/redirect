import storage


def get_storage_creator(config):
    spec = storage.mysql_spec_config(config)
    maker = storage.get_session_maker(spec)
    create_storage = storage.SessionContextFactory(maker)
    return create_storage
