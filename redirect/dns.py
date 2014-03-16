from boto.route53.record import ResourceRecordSets
import boto.route53


class Dns:

    def __init__(self, aws_access_key_id, aws_secret_access_key, hosted_zone_id):
        self.aws_access_key_id = aws_access_key_id
        self.aws_secret_access_key = aws_secret_access_key
        self.hosted_zone_id = hosted_zone_id

    def create_records(self, username, ip, port, domain):

        conn = boto.connect_route53(self.aws_access_key_id, self.aws_secret_access_key)
        changes = ResourceRecordSets(conn, self.hosted_zone_id)

        change = changes.add_change('CREATE', '{0}.{1}.'.format(username, domain), 'CNAME')
        change.add_value(domain)

        change = changes.add_change('CREATE', 'device.{0}.{1}.'.format(username, domain), 'A')
        change.add_value(ip)

        change = changes.add_change('CREATE', '_owncloud._http._tcp.{0}.{1}.'.format(username, domain), 'SRV')
        change.add_value('0 0 {0} device.{1}.{2}.'.format(port, username, domain))

        changes.commit()

    def update_records(self, username, ip, port, domain):

        conn = boto.connect_route53(self.aws_access_key_id, self.aws_secret_access_key)
        changes = ResourceRecordSets(conn, self.hosted_zone_id)

        change = changes.add_change('UPSERT', 'device.{0}.{1}.'.format(username, domain), 'A')
        change.add_value(ip)

        change = changes.add_change('UPSERT', '_owncloud._http._tcp.{0}.{1}.'.format(username, domain), 'SRV')
        change.add_value('0 0 {0} device.{1}.{2}.'.format(port, username, domain))

        changes.commit()

    def delete_records(self, username, ip, port, domain):

        conn = boto.connect_route53(self.aws_access_key_id, self.aws_secret_access_key)
        changes = ResourceRecordSets(conn, self.hosted_zone_id)

        change = changes.add_change('DELETE', '{0}.{1}.'.format(username, domain), 'CNAME')
        change.add_value(domain)
        change = changes.add_change('DELETE', 'device.{0}.{1}.'.format(username, domain), 'A')
        change.add_value(ip)
        change = changes.add_change('DELETE', '_owncloud._http._tcp.{0}.{1}.'.format(username, domain), 'SRV')
        change.add_value('0 0 {0} device.{1}.{2}.'.format(port, username, domain))

        changes.commit()