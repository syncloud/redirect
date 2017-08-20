from boto.route53.record import ResourceRecordSets
import boto.route53
from IPy import IP


class Dns:

    def __init__(self, aws_access_key_id, aws_secret_access_key, hosted_zone_id):
        self.aws_access_key_id = aws_access_key_id
        self.aws_secret_access_key = aws_secret_access_key
        self.hosted_zone_id = hosted_zone_id

    def a_change(self, changes, ip, full_domain, change_action, ip_version):
        change_type = 'A'
        if ip_version == 6:
            change_type = 'AAAA'

        change = changes.add_change(change_action, full_domain, change_type)
        change.add_value(ip)

    def spf_change(self, changes, ip, full_domain, change_action, ip_version):
        change = changes.add_change(change_action, full_domain, 'SPF')
        spf_value = '"v=spf1 ip{0}:{1}-all"'.format(ip_version, ip)
        change.add_value(spf_value)

    def mx_change(self, changes, full_domain, change_action):
        change = changes.add_change(change_action, full_domain, 'MX')
        change.add_value('1 {0}'.format(full_domain))

    def update_domain(self, main_domain, domain):
        self.__action_domain(main_domain, domain, 'UPSERT')

    def delete_domain(self, main_domain, domain):
        self.__action_domain(main_domain, domain, 'DELETE')

    def __action_domain(self, main_domain, domain, action):

        ip = domain.dns_ip()
        if ip is None:
            return

        conn = boto.connect_route53(self.aws_access_key_id, self.aws_secret_access_key)
        changes = ResourceRecordSets(conn, self.hosted_zone_id)

        full_domain = domain.dns_name(main_domain)
        ip_version = IP(ip).version()
        self.a_change(changes, ip, full_domain, action, ip_version)
        self.a_change(changes, ip, '*.{0}'.format(full_domain), action, ip_version)
        self.mx_change(changes, full_domain, action)
        self.spf_change(changes, ip, full_domain, action, ip_version)

        changes.commit()
