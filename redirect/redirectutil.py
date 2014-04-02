import hashlib

def hash(plain_password):
    return hashlib.sha256(plain_password).hexdigest()


