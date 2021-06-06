package cli

// TODO: port dns clean

/*
sys.path.append(normpath(join(dirname(__file__), '..')))
from redirect import ioc
import argparse

if __name__=='__main__':
    parser = argparse.ArgumentParser(description='This is redirect unsubscribing email tool', usage='%(prog)s [options]')
    parser.add_argument('date', help='data before cleanup')
    parser.add_argument('limit', help='how many to cleanup')
    args = parser.parse_args()

    date = args.date
    limit = args.limit

    manager = ioc.manager()
    create_storage = manager.create_storage
    dns = manager.dns
    with create_storage() as storage:
        for domain in storage.get_domains_last_updated_before(date, limit):
            print('ip: {0}, last update: {1}'.format(domain.ip, domain.last_update))
            dns.delete_domain(manager.main_domain, domain)
            domain.ip = None

*/
