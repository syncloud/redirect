import unittest
import ConfigParser
import os


class TestConfig(unittest.TestCase):

    def test_config(self):

        config = ConfigParser.ConfigParser()
        config.read(os.path.dirname(__file__) + '/../config.cfg.dist')
        self.assertEquals(config.get('mysql', 'host'), 'localhost')

