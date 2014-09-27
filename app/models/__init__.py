"""
This file contains the data models for passing data too and from the datastore
https://developers.google.com/appengine/docs/python/ndb/
https://cloud.google.com/appengine/docs/python/ndb/properties
"""

from google.appengine.ext import ndb


# Cannot use keys: "key", "id", "parent", or "namespace"
class Post(ndb.Model):
    tags = ndb.StringProperty(repeated=True)
    link = ndb.StringProperty()
    date = ndb.StringProperty()
    guid = ndb.StringProperty()  # also the ID
    title = ndb.StringProperty()
    media = ndb.JsonProperty()
    creator = ndb.JsonProperty()
    keys = ndb.KeyProperty(repeated=True)


class Img(ndb.Model):
    url = ndb.StringProperty()  # also the ID
    title = ndb.TextProperty()
    rating = ndb.StringProperty()
    category = ndb.StringProperty(repeated=True)
    is_valid = ndb.BooleanProperty(default=True)
