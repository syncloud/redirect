import configparser
from os.path import join, exists


def read_configs(filenames):
    missing_filenames = [f for f in filenames if not exists(f)]
    if missing_filenames:
        print('Missing configuration files: '+str(missing_filenames))

    config = configparser.ConfigParser()
    config.read(filenames)
    return config


def read_redirect_configs(config_dir):
    config = read_configs([join(config_dir, f) for f in ['config.cfg', 'secret.cfg']])
    return config
