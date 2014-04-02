
class RestException(Exception):
    status_code = 400

    def __init__(self, message, status_code=None):
        Exception.__init__(self)
        self.message = message
        if status_code is not None:
            self.status_code = status_code

def bad_request(message):
    return RestException(message)

def conflict(message):
    return RestException(message, status_code=409)