from redirect.models import Domain

def test_ipv6():
    dns = Domain('user_domain', '0:0:0:0:0', 'device_name', 'device_title', 'update_token')
    ipv6 = "2001:0db8:85a3:0000:0000:8a2e:0370:7334"
    dns.ipv6 = ipv6
    assert dns.dns_ipv6() == ipv6
    assert dns.dns_ipv4() is None

def test_ipv6():
    dns = Domain('user_domain', '0:0:0:0:0', 'device_name', 'device_title', 'update_token')
    ipv4 = "192.168.0.1"
    dns.ip = ipv4
    assert dns.dns_ipv6() is None
    assert dns.dns_ipv4() == ipv4


def test_access_ip_external():
    dns = Domain('user_domain', '0:0:0:0:0', 'device_name', 'device_title', 'update_token')
    dns.ip = "192.168.0.1"
    dns.local_ip = "192.168.0.2"
    dns.map_local_address = False
    assert dns.access_ip() == "192.168.0.1"


def test_access_ip_local():
    dns = Domain('user_domain', '0:0:0:0:0', 'device_name', 'device_title', 'update_token')
    dns.ip = "192.168.0.1"
    dns.local_ip = "192.168.0.2"
    dns.map_local_address = True
    assert dns.access_ip() == "192.168.0.2"

def test_has_dns_ip():
    dns = Domain('user_domain', '0:0:0:0:0', 'device_name', 'device_title', 'update_token')
    dns.ip = None
    dns.local_ip = None
    dns.ipv6 = None
    assert not dns.has_dns_ip()
