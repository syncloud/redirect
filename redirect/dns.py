from boto.route53.record import ResourceRecordSets
import boto.route53


class Dns:

    def __init__(self, aws_access_key_id, aws_secret_access_key, hosted_zone_id):
        self.aws_access_key_id = aws_access_key_id
        self.aws_secret_access_key = aws_secret_access_key
        self.hosted_zone_id = hosted_zone_id

    def service_change(self, changes, main_domain, user_domain, change_type, service):
            service_change = '{0}.{1}.{2}.'.format(service.type, user_domain, main_domain)
            service_value = '0 0 {0} device.{1}.{2}.'.format(service.port, user_domain, main_domain)
            change = changes.add_change(change_type, service_change, 'SRV')
            change.add_value(service_value)

    def services_change(self, changes, main_domain, domain, change_type, services):
        for s in services:
            self.service_change(changes, main_domain, domain.user_domain, change_type, s)

    def a_change(self, changes, main_domain, domain, change_type):
            change = changes.add_change(change_type, 'device.{0}.{1}.'.format(domain.user_domain, main_domain), 'A')
            change.add_value(domain.ip)

    def cname_change(self, changes, main_domain, domain, change_type):
        change = changes.add_change(change_type, '{0}.{1}.'.format(domain.user_domain, main_domain), 'CNAME')
        change.add_value(main_domain)

    def new_domain(self, main_domain, domain):
        conn = boto.connect_route53(self.aws_access_key_id, self.aws_secret_access_key)
        changes = ResourceRecordSets(conn, self.hosted_zone_id)

        self.cname_change(changes, main_domain, domain, 'CREATE')
        self.a_change(changes, main_domain, domain, 'CREATE')
        self.services_change(changes, main_domain, domain, 'CREATE', domain.services)

        changes.commit()

    def update_domain(self, main_domain, domain, update_ip=False, added=[], removed=[]):
        conn = boto.connect_route53(self.aws_access_key_id, self.aws_secret_access_key)
        changes = ResourceRecordSets(conn, self.hosted_zone_id)

        if update_ip:
            self.a_change(changes, main_domain, domain, 'UPSERT')
        self.services_change(changes, main_domain, domain, 'DELETE', removed)
        self.services_change(changes, main_domain, domain, 'UPSERT', added)

        changes.commit()

    def delete_domain(self, main_domain, domain):
        conn = boto.connect_route53(self.aws_access_key_id, self.aws_secret_access_key)
        changes = ResourceRecordSets(conn, self.hosted_zone_id)

        self.cname_change(changes, main_domain, domain, 'DELETE')
        self.a_change(changes, main_domain, domain, 'DELETE')
        self.services_change(changes, main_domain, domain, 'DELETE', domain.services)

        changes.commit()