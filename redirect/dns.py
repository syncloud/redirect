import logging
from boto.route53.record import ResourceRecordSets
import boto.route53


class Dns:

    def __init__(self, aws_access_key_id, aws_secret_access_key, hosted_zone_id):
        self.aws_access_key_id = aws_access_key_id
        self.aws_secret_access_key = aws_secret_access_key
        self.hosted_zone_id = hosted_zone_id

    def service_change(self, changes, main_domain, change_type, service):
            change = changes.add_change(change_type, service.dns_name(main_domain), 'SRV')
            change.add_value(service.dns_value(main_domain))

    def services_change(self, changes, main_domain, change_type, services):
        for s in services:
            self.service_change(changes, main_domain, change_type, s)

    def a_change(self, changes, main_domain, domain, change_type):
        change = changes.add_change(change_type, domain.dns_name(main_domain), 'A')
        change.add_value(domain.ip)

    def cname_change(self, changes, main_domain, domain, change_type):
        change = changes.add_change(change_type, domain.dns_name(main_domain), 'CNAME')
        change.add_value(main_domain)

    def new_domain(self, main_domain, domain):
        conn = boto.connect_route53(self.aws_access_key_id, self.aws_secret_access_key)
        changes = ResourceRecordSets(conn, self.hosted_zone_id)

        self.cname_change(changes, main_domain, domain, 'UPSERT')
        self.a_change(changes, main_domain, domain, 'UPSERT')
        self.services_change(changes, main_domain, 'UPSERT', domain.services)

        changes.commit()

    def update_domain(self, main_domain, domain, update_ip=False, added=[], removed=[]):

        if not update_ip and len(added) == 0 and len(removed) == 0:
            return

        conn = boto.connect_route53(self.aws_access_key_id, self.aws_secret_access_key)
        changes = ResourceRecordSets(conn, self.hosted_zone_id)

        if update_ip:
            self.a_change(changes, main_domain, domain, 'UPSERT')

        existing = [s for s in removed if self.service_exists(conn, s, main_domain)]
        self.services_change(changes, main_domain, 'DELETE', existing)

        self.services_change(changes, main_domain, 'UPSERT', added)

        try:
            changes.commit()
        except Exception, e:
            logging.error("added: {}".format(added))
            logging.error("removed: {}".format(removed))
            raise e

    def service_exists(self, conn, service, main_domain):
        zone = conn.get_zone(main_domain)
        found = zone.find_records(service.dns_name(main_domain), 'SRV')
        if not found:
            return False
        if len(found.resource_records) > 0:
            return found.resource_records[0] == service.dns_value(main_domain)
        else:
            return False


    def delete_domain(self, main_domain, domain):
        conn = boto.connect_route53(self.aws_access_key_id, self.aws_secret_access_key)
        changes = ResourceRecordSets(conn, self.hosted_zone_id)
        zone = conn.get_zone(main_domain)

        if zone.find_records(domain.dns_name(main_domain)):
            self.cname_change(changes, main_domain, domain, 'DELETE')

        if zone.find_records(domain.dns_name(main_domain)):
            self.a_change(changes, main_domain, domain, 'DELETE')

        existing = [s for s in domain.services if zone.find_records(s.dns_name(main_domain), 'SRV')]
        self.services_change(changes, main_domain, 'DELETE', existing)

        changes.commit()