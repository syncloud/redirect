from os.path import join, exists, dirname
import ConfigParser

def read_configs(filenames):
    missing_filenames = [f for f in filenames if not exists(f)]
    if missing_filenames:
        print('Missing configuration files: '+str(missing_filenames))

    config = ConfigParser.ConfigParser()
    config.read(filenames)
    return config

def read_redirect_configs():
    file_dirname = dirname(__file__)
    config = read_configs([join(file_dirname, f) for f in ['config.cfg', 'test_secret.cfg']])
    return config
