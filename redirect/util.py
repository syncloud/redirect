import uuid

def create_token():
    return unicode(uuid.uuid4().hex)

