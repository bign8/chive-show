from google.appengine.ext import ndb
from app.models import Post, Img
import random


def post_random(count):
    """ grab a subset of posts for viewing """

    # TODO: mark content as viewed by this wb when sent full list
    # TODO: keep track of sessions / cookies
    # TODO: implement user preferences

    # Don't get massive amounts of random posts
    if count > 30:
        return {'status':'error','code':500,'data':'Requested too many posts'}

    # Thanks: http://stackoverflow.com/a/21650400
    query = Post.query()
    all_keys = query.fetch(keys_only=True)
    if len(all_keys) < count:
        return {'status':'error','code':500,'data':'Basically empty datastore'}

    # Get 10 random posts
    list_keys = random.sample(all_keys, count)
    posts = ndb.get_multi(list_keys)  # get all the keys at once

    # Flattem images and query them all at once
    image_keys = [key for post in posts for key in post.keys]
    images = ndb.get_multi(image_keys)  # get all images at once

    # Re-populate posts with images
    start = 0
    for post in posts:
        end = start + len(post.keys)
        post.media = [img.to_dict() for img in images[start:end]]
        start = end

    # Convert objects to dicts
    exclude = ['keys', 'urlsafe']
    data = [post.to_dict(exclude=exclude) for post in posts]

    return {'status':'success','code':200,'data':data}


def image_info(urlsafe=None):
    """ Update image metadata (server never loads images) """
    try:
        img = Img.from_urlsafe(urlsafe)
        return img.to_dict()
    except:
        return {'status':'error','code':500,'data':'Cannot retrieve image'}

def tags():
    """ returns list of all tags """
    tags_query = Post.query(projection=['tags'], distinct=True)
    tags = [x.tags[0] for x in tags_query.fetch()]
    return {'status':'success','code':200,'data':tags}
