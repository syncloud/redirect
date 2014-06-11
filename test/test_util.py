import unittest
from redirect.util import hash, get_second_level_domain

class TestHash(unittest.TestCase):

    def test_empty(self):
        h = hash('non empty string')
        self.assertIsNotNone(h)
        self.assertIsNot('', h)

    def test_equal_input(self):
        h1 = hash('some string')
        h2 = hash('some string')
        self.assertEqual(h1, h2)

    def test_not_equal_input(self):
        h1 = hash('some string')
        h2 = hash('some other string')
        self.assertNotEqual(h1, h2)

class TestGetSecondLevelDomain(unittest.TestCase):

    def test_simple(self):
        domain = 'syncloud.com'
        url = 'http://second.syncloud.com'
        second_level_domain_name = get_second_level_domain(url, domain)
        self.assertEquals('second', second_level_domain_name)

    def test_long_url(self):
        domain = 'syncloud.com'
        url = 'http://second.syncloud.com/some/really/long/url/with_param=value'
        second_level_domain_name = get_second_level_domain(url, domain)
        self.assertEquals('second', second_level_domain_name)

    def test_port(self):
        domain = 'syncloud.com'
        url = 'http://second.syncloud.com:10001/param=value'
        second_level_domain_name = get_second_level_domain(url, domain)
        self.assertEquals('second', second_level_domain_name)


if __name__ == '__main__':
    unittest.run()