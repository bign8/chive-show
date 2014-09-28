"""
This file contains the data models for passing data too and from the datastore
https://developers.google.com/appengine/docs/python/ndb/
https://cloud.google.com/appengine/docs/python/ndb/properties
"""

from google.appengine.ext import ndb

class _DB_Helpers(ndb.Model):

    @classmethod
    def from_urlsafe(cls, urlsafe):
        """ Retrive instance of class that corresponds to urlsafe key """
        try:
            key = ndb.Key(urlsafe=urlsafe)
            obj = key.get() if key.kind() == cls.__name__ else None
            return obj
        except:
            raise TypeError('Invalid urlsafe Key')

    def to_dict(self, **args):
        """ Extend to_dict to append urlsafe key from db """
        data = super(_DB_Helpers, self).to_dict(**args)
        if 'urlsafe' not in args.get('exclude', []):
            data['urlsafe'] = self.key.urlsafe()
        return data


# Cannot use keys: "key", "id", "parent", or "namespace"
class Post(_DB_Helpers):
    tags = ndb.StringProperty(repeated=True)
    link = ndb.StringProperty()
    date = ndb.StringProperty()
    guid = ndb.StringProperty()  # also the ID
    title = ndb.StringProperty()
    media = ndb.JsonProperty()
    creator = ndb.JsonProperty()
    keys = ndb.KeyProperty(repeated=True)


class Img(_DB_Helpers):
    url = ndb.StringProperty()  # also the ID
    title = ndb.TextProperty()
    rating = ndb.StringProperty()
    category = ndb.StringProperty(repeated=True)
    is_valid = ndb.BooleanProperty(default=True)
