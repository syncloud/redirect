import hashlib
import uuid
from urlparse import urlparse

def hash(plain_password):
    return hashlib.sha256(plain_password).hexdigest()

def create_token():
    return unicode(uuid.uuid4().hex)

