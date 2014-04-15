from __future__ import absolute_import
import socket
import urllib
import unittest
import dns.resolver as resolver
import time
import ConfigParser
import os
import json

config = ConfigParser.ConfigParser()
config.read(os.path.dirname(__file__) + '/test_config.cfg')

test_url = config.get('full_cycle', 'test_url')
user = 'user1'
password = 'pass123'
email = user + '@example.com'


class TestIntegrationCycle(unittest.TestCase):
    def test_full_cycle(self):

        # Create (auto activate)
        response = urllib.urlopen(
            "http://{}/user/create?user_domain={}&password={}&email={}"
            .format(test_url, user, password, email))
        self.assertEquals(response.getcode(), 200, response.read())

        # Get token
        response = urllib.urlopen(
            "http://{}/user/get?email={}&password={}"
            .format(test_url, email, password))
        user_data = json.loads(response.read())
        token = user_data['update_token']
        self.assertTrue(token is not None, token)

        # Check DNS (not available yet)
        self.assertFalse(self.wait_for_dns(user + '.test.com', 'CNAME'))

        # Change IP/port (one)
        response = urllib.urlopen("http://{1}/domain/update?token={0}&ip=192.168.0.1&port=80".format(token, test_url))
        self.assertEquals(response.getcode(), 200, response.read())

        #  Validate DNS (one)
        self.assertEquals(
            self.wait_for_dns('device.{0}.test.com'.format(user), 'A', lambda v: v.address == '192.168.0.1').address,
            '192.168.0.1')
        srv = self.wait_for_dns('_owncloud._http._tcp.{0}.test.com'.format(user), 'SRV')
        self.assertEquals(srv.port, 80)
        self.assertEquals(srv.target.to_text(True), 'device.{0}.test.com'.format(user))

        # Change IP/port (two)
        response = urllib.urlopen("http://{1}/domain/update?token={0}&ip=192.168.0.2&port=81".format(token, test_url))
        self.assertEquals(response.getcode(), 200, response.read())

        #  Validate DNS (two)
        self.assertEquals(
            self.wait_for_dns('device.{0}.test.com'.format(user), 'A', lambda v: v.address == '192.168.0.2').address,
            '192.168.0.2')
        srv = self.wait_for_dns('_owncloud._http._tcp.{0}.test.com'.format(user), 'SRV')
        self.assertEquals(srv.port, 81)
        self.assertEquals(srv.target.to_text(True), 'device.{0}.test.com'.format(user))

        # Remove
        response = urllib.urlopen("http://{1}/user/delete?email={0}&password={}".format(email, password, test_url))
        self.assertEquals(response.getcode(), 200, response.read())

        # Check DNS (nothing)
        self.assertFalse(self.wait_for_dns(user + '.test.com', 'CNAME', lambda v: not v))

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