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
import logging


class TestDns(unittest.TestCase):

    def setUp(self):
        self.config = ConfigParser.ConfigParser()
        self.config.read(os.path.dirname(__file__) + '/config.cfg')

        self.main_domain = self.config.get('full_cycle', 'domain')
        self.user_domain = 'user11'

        self.dns = Dns(
            self.config.get('aws', 'access_key_id'),
            self.config.get('aws', 'secret_access_key'),
            self.config.get('aws', 'hosted_zone_id'))

        logging.getLogger('boto').setLevel(logging.DEBUG)

    def test(self):
        update_token = create_token()

        full_domain_name = 'name1.{0}.{1}'.format(self.user_domain, self.main_domain)
        original_ip = self.query_dns(full_domain_name, 'A')

        domain = Domain(self.user_domain, '00:00:00:00:00:00', 'some-device', 'Some Device', update_token)
        domain.ip = '192.168.0.1'
        self.dns.update_domain(self.main_domain, domain)
        full_domain_name = 'name1.{0}.{1}'.format(self.user_domain, self.main_domain)
        self.validate_dns('192.168.0.1', full_domain_name)

        domain.ip = '192.168.0.2'
        self.dns.update_domain(self.main_domain, domain)
        full_domain_name = 'name1.{0}.{1}'.format(self.user_domain, self.main_domain)
        self.validate_dns('192.168.0.2', full_domain_name)

        self.dns.delete_domain(self.main_domain, domain)
        full_domain_name = 'name1.{0}.{1}'.format(self.user_domain, self.main_domain)
        self.validate_dns(original_ip.address, full_domain_name)

    def validate_dns(self, ip, full_domain_name):
        record = self.wait_for_dns(full_domain_name, 'A', ip)
        self.assertEquals(record.address, ip)

    def wait_for_dns(self, name, name_type, ip):

        stop = 50
        for i in range(0, stop):
            value = self.query_dns(name, name_type)
            if value.address == ip:
                print "found: {0}/{1} => {2}".format(name, name_type, value)
                return value
            time.sleep(1)
            print "waiting for {0}, {1}, {2}, got: {3} ... {4}/{5}".format(name, name_type, ip, value.address, i, stop)

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