from __future__ import absolute_import
import socket
import unittest
import dns.resolver as resolver
import time
import ConfigParser
import os
from redirect.dns import Dns

config = ConfigParser.ConfigParser()
config.read(os.path.dirname(__file__) + '/test_config.cfg')

domain = config.get('full_cycle', 'domain')
user_domain = 'user7'


class TestDns(unittest.TestCase):

    def test(self):

        dns = Dns(
            config.get('aws', 'access_key_id'),
            config.get('aws', 'secret_access_key'),
            config.get('aws', 'hosted_zone_id'))

        dns.create_records(user_domain, '192.168.0.1', '80', domain)
        self.validate_dns('192.168.0.1', 80)

        dns.update_records(user_domain, '192.168.0.2', '81', domain)
        self.validate_dns('192.168.0.2', 81)

        dns.delete_records(user_domain, '192.168.0.2', '81', domain)
        self.assertFalse(self.wait_for_dns('{0}.{1}'.format(user_domain, domain), 'CNAME', lambda v: not v))

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