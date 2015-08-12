from __future__ import absolute_import
import socket
import unittest
import dns.resolver as resolver
import time
import ConfigParser
import os

from redirect.dns import Dns
from redirect.models import Domain, new_service
from redirect.util import create_token


class TestDns(unittest.TestCase):

    def setUp(self):
        config = ConfigParser.ConfigParser()
        config.read(os.path.dirname(__file__) + '/config.cfg')

        self.main_domain = config.get('full_cycle', 'domain')
        self.user_domain = 'user11'

        self.dns = Dns(
            config.get('aws', 'access_key_id'),
            config.get('aws', 'secret_access_key'),
            config.get('aws', 'hosted_zone_id'))

    def test(self):
        update_token = create_token()

        domain = Domain(self.user_domain, '00:00:00:00:00:00', 'some-device', 'Some Device', update_token)
        domain.ip = '192.168.0.1'

        self.dns.update_domain(self.main_domain, domain)

        service_80 = new_service('ssh', '_ssh._tcp', 80)
        service_80.domain = domain
        self.dns.update_domain(self.main_domain, domain)
        self.validate_dns('192.168.0.1')

        domain.ip = '192.168.0.2'
        service_81 = new_service('web', '_www._tcp', 81)
        service_81.domain = domain
        self.dns.update_domain(self.main_domain, domain)
        self.validate_dns('192.168.0.2')

        domain.ip = '192.168.0.3'
        service_82 = new_service('web1', '_www1._tcp', 82)
        service_82.domain = domain
        self.dns.update_domain(self.main_domain, domain)
        self.validate_dns('192.168.0.3')

        self.dns.update_domain(self.main_domain, domain)

        self.dns.delete_domain(self.main_domain, domain)
        self.assertFalse(self.wait_for_dns('{0}.{1}'.format(self.main_domain, domain), 'CNAME', lambda v: not v))

    def validate_dns(self, ip):
        full_domain_name = 'device.{0}.{1}'.format(self.user_domain, self.main_domain)
        record = self.wait_for_dns(full_domain_name, 'A', lambda v: v and v.address == ip)
        self.assertEquals(record.address, ip)

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
        res.nameservers = [socket.gethostbyname(self.config.get('full_cycle', 'route_53_server'))]

        try:
            response = res.query(name, name_type)
            for rdata in response:
                return rdata
        except Exception:
            return None