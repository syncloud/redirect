from __future__ import absolute_import
import socket
import unittest
import dns.resolver as resolver
import time
import ConfigParser
import os
from redirect.dns import Dns
from redirect.models import Domain, Service, new_service

config = ConfigParser.ConfigParser()
config.read(os.path.dirname(__file__) + '/test.config.cfg')

main_domain = config.get('full_cycle', 'domain')
user_domain = 'user9'


class TestDns(unittest.TestCase):

    def test(self):

        dns = Dns(
            config.get('aws', 'access_key_id'),
            config.get('aws', 'secret_access_key'),
            config.get('aws', 'hosted_zone_id'))

        domain = Domain(user_domain, '192.168.0.1')
        dns.new_domain(main_domain, domain)

        service_80 = new_service('web', '_http._tcp', 80)

        dns.update_domain(main_domain, domain, None, [service_80], [])
        self.validate_dns('192.168.0.1', 80)

        domain.ip = '192.168.0.2'
        service_81 = new_service('web', '_http._tcp', 81)
        dns.update_domain(main_domain, domain, True, [service_81], [service_80])
        self.validate_dns('192.168.0.2', 81)

        dns.delete_domain(main_domain, domain)
        self.assertFalse(self.wait_for_dns('{0}.{1}'.format(main_domain, domain), 'CNAME', lambda v: not v))

    def validate_dns(self, ip, port):
        self.assertEquals(
            self.wait_for_dns(
                'device.{0}.{1}'.format(user_domain, main_domain),
                'A',
                lambda v: v and v.address == ip).address,
            ip)
        srv = self.wait_for_dns('_http._tcp.{0}.{1}'.format(user_domain, main_domain), 'SRV')
        self.assertEquals(srv.port, port)
        self.assertEquals(srv.target.to_text(True), 'device.{0}.{1}'.format(user_domain, main_domain))

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