from __future__ import absolute_import
import socket
import unittest
import dns.resolver as resolver
import time
import ConfigParser
import os

from redirect.dns import Dns
from redirect.models import Domain, new_service


config = ConfigParser.ConfigParser()
config.read(os.path.dirname(__file__) + '/config.cfg')

main_domain = config.get('full_cycle', 'domain')
user_domain = 'user11'


class TestDns(unittest.TestCase):

    def test(self):

        dns = Dns(
            config.get('aws', 'access_key_id'),
            config.get('aws', 'secret_access_key'),
            config.get('aws', 'hosted_zone_id'))

        domain = Domain(user_domain, '192.168.0.1')
        dns.new_domain(main_domain, domain)

        service_80 = new_service('ssh', '_ssh._tcp', 80)
        service_80.domain = Domain(user_domain)

        dns.update_domain(main_domain, domain, None, [service_80], [])
        self.validate_dns('192.168.0.1', 80, '_ssh._tcp')

        domain.ip = '192.168.0.2'
        service_81 = new_service('web', '_www._tcp', 81)
        service_81.domain = Domain(user_domain)
        dns.update_domain(main_domain, domain, True, [service_81], [service_80])
        self.validate_dns('192.168.0.2', 81, '_www._tcp')

        domain.ip = '192.168.0.3'
        service_82 = new_service('web1', '_www1._tcp', 82)
        service_82.domain = Domain(user_domain)
        dns.update_domain(main_domain, domain, True, [service_82], [])
        self.validate_dns('192.168.0.3', 82, '_www1._tcp')

        dns.update_domain(main_domain, domain, True, [service_82], [])

        dns.delete_domain(main_domain, domain)
        self.assertFalse(self.wait_for_dns('{0}.{1}'.format(main_domain, domain), 'CNAME', lambda v: not v))

    def validate_dns(self, ip, port, srv_type):
        self.assertEquals(
            self.wait_for_dns(
                'device.{0}.{1}'.format(user_domain, main_domain),
                'A',
                lambda v: v and v.address == ip).address,
            ip)
        srv = self.wait_for_dns('{0}.{1}.{2}'.format(srv_type, user_domain, main_domain), 'SRV')
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