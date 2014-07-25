import datetime
from flask.json import JSONEncoder

class CustomJSONEncoder(JSONEncoder):

    def default(self, obj):
        try:
            if isinstance(obj, datetime.datetime):
                return str(obj)
            iterable = iter(obj)
        except TypeError, e:
            pass
        else:
            return list(iterable)
        return JSONEncoder.default(self, obj)