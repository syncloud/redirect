import config
import db_helper
from dns import Dns

from boto.route53.record import ResourceRecordSets
import boto.route53


if __name__=='__main__':
    the_config = config.read_redirect_configs()

    main_domain = the_config.get('redirect', 'domain')

    dns = Dns(
        the_config.get('aws', 'access_key_id'),
        the_config.get('aws', 'secret_access_key'),
        the_config.get('aws', 'hosted_zone_id'))

    create_storage = db_helper.get_storage_creator(the_config)

    with create_storage() as storage:
        for domain in storage.domains_iterate():
            aws_access_key_id = the_config.get('aws', 'access_key_id')
            aws_secret_access_key = the_config.get('aws', 'secret_access_key')
            hosted_zone_id = the_config.get('aws', 'hosted_zone_id')

            conn = boto.connect_route53(aws_access_key_id, aws_secret_access_key)
            changes = ResourceRecordSets(conn, hosted_zone_id)

            old_cname = '{0}.{1}.'.format(domain.user_domain, main_domain)
            changes.add_change('DELETE', old_cname, 'CNAME')

            old_a_record = 'device.{0}.{1}.'.format(domain.user_domain, main_domain)
            changes.add_change('DELETE', old_a_record, 'A')

            changes.add_change('UPSERT', old_cname, 'A')

            print('Removing CNAME {}'.format(old_cname))
            print('Removing A {}'.format(old_a_record))
            print('Adding A {}'.format(old_cname))

            changes.commit()