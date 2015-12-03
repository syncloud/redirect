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

    def update_domain(self, main_domain, domain):
        conn = boto.connect_route53(self.aws_access_key_id, self.aws_secret_access_key)
        changes = ResourceRecordSets(conn, self.hosted_zone_id)

        ip = domain.dns_ip()
        full_domain = domain.dns_name(main_domain)

        self.a_change(changes, ip, full_domain, 'UPSERT')
        self.a_change(changes, ip, '*.{0}'.format(full_domain), 'UPSERT')

        changes.commit()

    def delete_domain(self, main_domain, domain):
        conn = boto.connect_route53(self.aws_access_key_id, self.aws_secret_access_key)
        zone = conn.get_zone(main_domain)

        full_domain = domain.dns_name(main_domain)
        if zone.find_records(full_domain, 'A'):
            zone.delete_a(full_domain)

        wildcard_domain = '*.{0}'.format(full_domain)
        if zone.find_records(wildcard_domain, 'A'):
            zone.delete_a(wildcard_domain)
