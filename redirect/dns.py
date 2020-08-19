from boto.route53.record import ResourceRecordSets
import boto.route53


class Dns:

    def __init__(self, aws_access_key_id, aws_secret_access_key, hosted_zone_id, statsd_client):
        self.aws_access_key_id = aws_access_key_id
        self.aws_secret_access_key = aws_secret_access_key
        self.hosted_zone_id = hosted_zone_id
        self.statsd_client = statsd_client
        self.default_ipv4 = '127.0.0.1'
        self.default_spf = '"v=spf1 -all"'
        self.default_ipv6 = 'fe80::'
        self.default_mx = '1 mx'
        self.default_dkim = 'none'

    def a_change(self, changes, ip, full_domain, action):
        change = changes.add_change(action, full_domain, "A")
        change.add_value(ip)

    def aaaa_change(self, changes, ip, full_domain, action):
        change = changes.add_change(action, full_domain, "AAAA")
        change.add_value(ip)

    def spf_change(self, changes, full_domain, spf, type, action):
        change = changes.add_change(action, full_domain, type)
        change.add_value(spf)

    def mx_change(self, changes, full_domain, mx, action):
        change = changes.add_change(action, full_domain, 'MX')
        change.add_value(mx)

    def dkim_change(self, changes, full_domain, dkim, action):
        name = 'mail._domainkey.{0}'.format(full_domain)
        dkim_value = '"v=DKIM1; k=rsa; p={0}"'.format(dkim)
        change = changes.add_change(action, name, 'TXT')
        change.add_value(dkim_value)
        
    def delete_domain(self, main_domain, domain):
        full_domain = domain.dns_name(main_domain)
        self.__action_domain(main_domain, full_domain,self.default_ipv4, self.default_ipv6, self.default_dkim, self.default_spf, self.default_mx, 'UPSERT')
        self.__action_domain(main_domain, full_domain,self.default_ipv4, self.default_ipv6, self.default_dkim, self.default_spf, self.default_mx, 'DELETE')

    def __action_domain(self, main_domain, full_domain, ipv4, ipv6, dkim, spf, mx, action):

        self.statsd_client.incr('dns.ip.connect')

        conn = boto.connect_route53(self.aws_access_key_id, self.aws_secret_access_key)
        changes = ResourceRecordSets(conn, self.hosted_zone_id)
        
        if ipv6:
            self.aaaa_change(changes, ipv6, full_domain, action)
            self.aaaa_change(changes, ipv6, '*.{0}'.format(full_domain), action)
        if ipv4:
            self.a_change(changes, ipv4, full_domain, action)
            self.a_change(changes, ipv4, '*.{0}'.format(full_domain), action)
        if dkim:
            self.dkim_change(changes, full_domain, dkim, action)
        if mx:
            self.mx_change(changes, full_domain, mx, action)
        if spf:
            self.spf_change(changes, full_domain, spf, 'SPF', action)
            self.spf_change(changes, full_domain, spf, 'TXT', action)
        
        self.statsd_client.incr('dns.ip.commit')
        changes.commit()

