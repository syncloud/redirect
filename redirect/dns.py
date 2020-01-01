from boto.route53.record import ResourceRecordSets
import boto.route53


class Dns:

    def __init__(self, aws_access_key_id, aws_secret_access_key, hosted_zone_id, statsd_client):
        self.aws_access_key_id = aws_access_key_id
        self.aws_secret_access_key = aws_secret_access_key
        self.hosted_zone_id = hosted_zone_id
        self.statsd_client = statsd_client
        
    def a_change(self, changes, ip, full_domain, change_action):
        change = changes.add_change(change_action, full_domain, "A")
        change.add_value(ip)

    def aaaa_change(self, changes, ip, full_domain, change_action):
        change = changes.add_change(change_action, full_domain, "AAAA")
        change.add_value(ip)

    def spf_change(self, changes, full_domain, change_action):
        spf_value = '"v=spf1 a mx -all"'

        change = changes.add_change(change_action, full_domain, 'SPF')
        change.add_value(spf_value)

        change = changes.add_change(change_action, full_domain, 'TXT')
        change.add_value(spf_value)

    def mx_change(self, changes, full_domain, change_action):
        change = changes.add_change(change_action, full_domain, 'MX')
        change.add_value('1 {0}'.format(full_domain))

    def dkim_change(self, changes, full_domain, change_action, dkim_key):
        name = 'mail._domainkey.{0}'.format(full_domain)
        value = '"v=DKIM1; k=rsa; p={0}"'.format(dkim_key)
        change = changes.add_change(change_action, name, 'TXT')
        change.add_value(value)

    def update_domain(self, main_domain, domain):
        self.__action_domain(main_domain, domain, 'UPSERT')

    def delete_domain(self, main_domain, domain):
        self.__action_domain(main_domain, domain, 'UPSERT')
        self.__action_domain(main_domain, domain, 'DELETE')

    def __action_domain(self, main_domain, domain, action):

        if not domain.has_dns_ip():
            return

        self.statsd_client.incr('dns.ip.connect')

        conn = boto.connect_route53(self.aws_access_key_id, self.aws_secret_access_key)
        changes = ResourceRecordSets(conn, self.hosted_zone_id)

        full_domain = domain.dns_name(main_domain)
        
        ipv6 = domain.dns_ipv6()
        if ipv6:
            self.aaaa_change(changes, ipv6, full_domain, action)
            self.aaaa_change(changes, ipv6, '*.{0}'.format(full_domain), action)
  
        ipv4 = domain.dns_ipv4()
        if ipv4:
            self.a_change(changes, ipv4, full_domain, action)
            self.a_change(changes, ipv4, '*.{0}'.format(full_domain), action)

        dkim_key = domain.dkim_key
        if dkim_key:
            self.dkim_change(changes, full_domain, action, dkim_key)

        self.mx_change(changes, full_domain, action)

        self.spf_change(changes, full_domain, action)
        
        self.statsd_client.incr('dns.ip.commit')
        changes.commit()
