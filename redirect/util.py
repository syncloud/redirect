import hashlib
import uuid
from urlparse import urlparse

def hash(plain_password):
    return hashlib.sha256(plain_password).hexdigest()

def create_token():
    return unicode(uuid.uuid4().hex)

def get_second_level_domain(request_url, domain):
    domain_address = urlparse(request_url).netloc
    address_no_port = domain_address.split(':')[0]
    if not address_no_port.endswith(domain):
        return None
    no_domain = address_no_port.replace(domain, '')
    user_domain = no_domain.split('.')[0]
    if user_domain == '':
        return None
    return user_domain


