from __future__ import absolute_import
import socket
import urllib
import unittest
import dns.resolver as resolver
import time
import ConfigParser
import os


config = ConfigParser.ConfigParser()
config.read(os.path.dirname(__file__) + '/test.config.cfg')

test_url = config.get('full_cycle', 'test_url')
user = 'user1'


class TestIntegrationCycle(unittest.TestCase):
    def test_full_cycle(self):

        # Create
        params = urllib.urlencode({
            'username': user,
            'email': user + '@example.com',
            'password': 'pass123',
            'ip': '192.168.0.1',  # TODO: Not needed on registration
            'port': '80'  # TODO: Not needed on registration
        })
        response = urllib.urlopen("http://%s/create?%s" % (test_url, params))
        self.assertEquals(response.getcode(), 200, response.read())

        token = response.headers.get('Token')
        self.assertTrue(token is not None, response.read())

        # Check DNS (nothing)
        self.assertFalse(self.wait_for_dns(user + '.test.com', 'CNAME'))

        # Activate
        response = urllib.urlopen("http://{1}/activate?token={0}".format(token, test_url))
        self.assertEquals(response.getcode(), 200, response.read())

        # Check DNS (something)
        self.assertTrue(self.wait_for_dns(user + '.test.com', 'CNAME'))

        # Change IP
        self.change_ip(token)

        # Remove
        response = urllib.urlopen("http://{1}/delete?username={0}&password=pass123".format(user, test_url))
        self.assertEquals(response.getcode(), 200, response.read())

        # Check DNS (nothing)
        self.assertFalse(self.wait_for_dns(user + '.test.com', 'CNAME', lambda v: not v))

    def change_ip(self, token):
        response = urllib.urlopen("http://{1}/update?token={0}&ip=192.168.0.1&port=80".format(token, test_url))
        self.assertEquals(response.getcode(), 200, response.read())
        self.assertEquals(
            self.wait_for_dns('device.{0}.test.com'.format(user), 'A', lambda v: v.address == '192.168.0.1').address,
            '192.168.0.1')
        srv = self.wait_for_dns('_owncloud._http._tcp.{0}.test.com'.format(user), 'SRV')
        self.assertEquals(srv.port, 80)
        self.assertEquals(srv.target.to_text(True), 'device.{0}.test.com'.format(user))

        response = urllib.urlopen("http://{1}/update?token={0}&ip=192.168.0.2&port=81".format(token, test_url))
        self.assertEquals(response.getcode(), 200, response.read())
        self.assertEquals(
            self.wait_for_dns('device.{0}.test.com'.format(user), 'A', lambda v: v.address == '192.168.0.2').address,
            '192.168.0.2')
        srv = self.wait_for_dns('_owncloud._http._tcp.{0}.test.com'.format(user), 'SRV')
        self.assertEquals(srv.port, 81)
        self.assertEquals(srv.target.to_text(True), 'device.{0}.test.com'.format(user))

    def wait_for_dns(self, name, name_type, condition=lambda v: v):

        stop = 50
        for i in range(0, stop):
            value = self.query_dns(name, name_type)
            if condition(value):
                print "found: {0}/{1} => {2}".format(name, name_type, value)
                return value
            time.sleep(1)
            print "waiting for {0}, {1} ... {2}/{3}".format(name, name_type, i, stop)

        return None

    def query_dns(self, name, name_type):

        res = resolver.Resolver()
        res.nameservers = [socket.gethostbyname(config.get('full_cycle', 'route_53_server'))]

        try:
            response = res.query(name, name_type)
            for rdata in response:
                return rdata
        except Exception:
            return None