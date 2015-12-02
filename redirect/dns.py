from boto.route53.record import ResourceRecordSets
import boto.route53


class Dns:

    def __init__(self, aws_access_key_id, aws_secret_access_key, hosted_zone_id):
        self.aws_access_key_id = aws_access_key_id
        self.aws_secret_access_key = aws_secret_access_key
        self.hosted_zone_id = hosted_zone_id

    def a_change(self, changes, ip, full_domain, change_type):
        change = changes.add_change(change_type, full_domain, 'A')
        change.add_value(ip)

    def update_domain(self, ip, full_domain, subdomains):
        conn = boto.connect_route53(self.aws_access_key_id, self.aws_secret_access_key)
        changes = ResourceRecordSets(conn, self.hosted_zone_id)

        self.a_change(changes, ip, full_domain, 'UPSERT')
        for subdomain in subdomains:
            self.a_change(changes, ip, '{0}.{1}'.format(subdomain, full_domain), 'UPSERT')

        changes.commit()

    def delete_domain(self, main_domain, domain):
        conn = boto.connect_route53(self.aws_access_key_id, self.aws_secret_access_key)
        zone = conn.get_zone(main_domain)

        dns_name = domain.dns_name(main_domain)
        if zone.find_records(dns_name, 'A'):
            zone.delete_a(dns_name)