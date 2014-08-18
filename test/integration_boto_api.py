import ConfigParser
import os
import unittest
from boto.route53.zone import Zone
import boto.route53

config = ConfigParser.ConfigParser()
config.read(os.path.dirname(__file__) + '/config.cfg')

main_domain = config.get('full_cycle', 'domain')
user_domain = 'user11'


class TestDns(unittest.TestCase):

    def test(self):

        conn = boto.connect_route53(config.get('aws', 'access_key_id'), config.get('aws', 'secret_access_key'))
        # zone = Zone(conn, { 'config.get('aws', 'hosted_zone_id'))

        zone = conn.get_zone('syncloud.info')
        # status = zone.get_a('device.ribalkin.syncloud.info.')

        found = zone.find_records('_ssh._tcp.testdomain1.syncloud.info', 'SRV')
        found.resource_records[0]

        print found.resource_records[0]