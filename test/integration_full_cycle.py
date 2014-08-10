from __future__ import absolute_import
import socket
import requests
import unittest
import dns.resolver as resolver
import time
import ConfigParser
import os
import json

config = ConfigParser.ConfigParser()
config.read(os.path.dirname(__file__) + '/config.cfg')

domain = config.get('full_cycle', 'domain')
test_url = 'api.{0}'.format(domain)
user_domain = 'user8'
password = 'pass123'
email = user_domain + '@example.com'


class TestIntegrationCycle(unittest.TestCase):
    def test_full_cycle(self):

        # Create (auto activate)
        response = requests.post(
            "http://{}/user/create".format(test_url),
            {'user_domain': user_domain, 'password': password, 'email': email})
        self.assertEquals(response.status_code, 200, response.content)

        # Get token
        response = requests.get(
            "http://{}/user/get".format(test_url),
            params={'email': email, 'password': password})
        user_data = json.loads(response.content)
        token = user_data['update_token']
        self.assertTrue(token is not None, token)

        # Check DNS (not available yet)
        self.assertFalse(self.wait_for_dns('{0}.{1}'.format(user_domain, domain), 'CNAME'))

        # Change IP/port (one)
        response = requests.post(
            "http://{}/domain/update".format(test_url),
            {'token': token, 'ip': '192.168.0.1', 'port': 80})
        self.assertEquals(response.status_code, 200, response.content)

        #  Validate DNS (one)
        self.validate_dns('192.168.0.1', 80)

        # Change IP/port (two)
        response = requests.post(
            "http://{}/domain/update".format(test_url),
            {'token': token, 'ip': '192.168.0.2', 'port': 81})
        self.assertEquals(response.status_code, 200, response.content)

        #  Validate DNS (two)
        self.validate_dns('192.168.0.2', 81)

        # Remove
        response = requests.post(
            "http://{0}/user/delete".format(test_url),
            {'email': email, 'password': password})
        self.assertEquals(response.status_code, 200, response.content)

        # Check DNS (nothing)
        self.assertFalse(self.wait_for_dns(user_domain + '.test.com', 'CNAME', lambda v: not v))

    def validate_dns(self, ip, port):
        self.assertEquals(
            self.wait_for_dns(
                'device.{0}.{1}'.format(user_domain, domain),
                'A',
                lambda v: v and v.address == ip).address,
            ip)
        srv = self.wait_for_dns('_owncloud._http._tcp.{0}.{1}'.format(user_domain, domain), 'SRV')
        self.assertEquals(srv.port, port)
        self.assertEquals(srv.target.to_text(True), 'device.{0}.{1}'.format(user_domain, domain))

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