from boto.route53.record import ResourceRecordSets
import boto.route53


class Dns:

    def __init__(self, aws_access_key_id, aws_secret_access_key, hosted_zone_id):
        self.aws_access_key_id = aws_access_key_id
        self.aws_secret_access_key = aws_secret_access_key
        self.hosted_zone_id = hosted_zone_id

    def service_record(self, changes, main_domain, user_domain, record_type, s):
            service_change = '{0}.{1}.{2}.'.format(s.type, user_domain, main_domain)
            change = changes.add_change(record_type, service_change, 'SRV')
            change.add_value('0 0 {0} device.{1}.{2}.'.format(s.port, user_domain, main_domain))

    def ip_and_services(self, changes, main_domain, domain, update_ip, created, changed, removed):
        if update_ip:
            change = changes.add_change('UPSERT', 'device.{0}.{1}.'.format(domain.user_domain, main_domain), 'A')
            change.add_value(domain.ip)
        for s in removed:
            self.service_record(changes, main_domain, domain.user_domain, 'DELETE', s)
        for s in created:
            self.service_record(changes, main_domain, domain.user_domain, 'CREATE', s)
        for s in changed:
            self.service_record(changes, main_domain, domain.user_domain, 'UPSERT', s)

    def new_domain(self, main_domain, domain):

        conn = boto.connect_route53(self.aws_access_key_id, self.aws_secret_access_key)
        changes = ResourceRecordSets(conn, self.hosted_zone_id)

        change = changes.add_change('CREATE', '{0}.{1}.'.format(domain.user_domain, main_domain), 'CNAME')
        change.add_value(main_domain)

        self.ip_and_services(changes, main_domain, domain, update_ip=True, created=domain.services)

        changes.commit()


    def update_domain(self, main_domain, domain, update_ip=False, created=[], changed=[], removed=[]):
        conn = boto.connect_route53(self.aws_access_key_id, self.aws_secret_access_key)
        changes = ResourceRecordSets(conn, self.hosted_zone_id)

        self.ip_and_services(changes, main_domain, domain, update_ip, created, changed, removed)

        changes.commit()

    def create_records(self, user_domain, ip, port, domain):

        conn = boto.connect_route53(self.aws_access_key_id, self.aws_secret_access_key)
        changes = ResourceRecordSets(conn, self.hosted_zone_id)

        change = changes.add_change('CREATE', '{0}.{1}.'.format(user_domain, domain), 'CNAME')
        change.add_value(domain)

        change = changes.add_change('CREATE', 'device.{0}.{1}.'.format(user_domain, domain), 'A')
        change.add_value(ip)

        change = changes.add_change('CREATE', '_owncloud._http._tcp.{0}.{1}.'.format(user_domain, domain), 'SRV')
        change.add_value('0 0 {0} device.{1}.{2}.'.format(port, user_domain, domain))

        changes.commit()

    def update_records(self, user_domain, ip, port, domain):

        conn = boto.connect_route53(self.aws_access_key_id, self.aws_secret_access_key)
        changes = ResourceRecordSets(conn, self.hosted_zone_id)

        change = changes.add_change('UPSERT', 'device.{0}.{1}.'.format(user_domain, domain), 'A')
        change.add_value(ip)

        change = changes.add_change('UPSERT', '_owncloud._http._tcp.{0}.{1}.'.format(user_domain, domain), 'SRV')
        change.add_value('0 0 {0} device.{1}.{2}.'.format(port, user_domain, domain))

        changes.commit()

    def delete_records(self, user_domain, ip, port, domain):

        conn = boto.connect_route53(self.aws_access_key_id, self.aws_secret_access_key)
        changes = ResourceRecordSets(conn, self.hosted_zone_id)

        change = changes.add_change('DELETE', '{0}.{1}.'.format(user_domain, domain), 'CNAME')
        change.add_value(domain)
        change = changes.add_change('DELETE', 'device.{0}.{1}.'.format(user_domain, domain), 'A')
        change.add_value(ip)
        change = changes.add_change('DELETE', '_owncloud._http._tcp.{0}.{1}.'.format(user_domain, domain), 'SRV')
        change.add_value('0 0 {0} device.{1}.{2}.'.format(port, user_domain, domain))

        changes.commit()