from google.appengine.ext import ndb
from app.models import Post, Img
import random

from app.decorators import timer

@timer.Timer('Get Keys')
def get_keys():
    # Thanks: http://stackoverflow.com/a/21650400
    # TODO: store random value with hashes + poll posts less than a random value (limit 10)
    query = Post.query()
    # TODO store this list or count to memcache
    return query.fetch(keys_only=True)


@timer.Timer('Get Posts')
def get_posts(all_keys, count):
    list_keys = random.sample(all_keys, count)
    return ndb.get_multi(list_keys)  # get all the keys at once


@timer.Timer('Get Images for posts')
def get_post_images(posts):
    # Flattem images and query them all at once
    image_keys = [key for post in posts for key in post.keys]
    return ndb.get_multi(image_keys)  # get all images at once


@timer.Timer('Format Response')
def format_response(posts, images):
    # Re-populate posts with images
    start = 0
    for post in posts:
        end = start + len(post.keys)
        post.media = [img.to_dict() for img in images[start:end]]
        start = end

    # Convert objects to dicts
    exclude = ['keys', 'urlsafe']
    return [post.to_dict(exclude=exclude) for post in posts]


def post_random(count):
    """ grab a subset of posts for viewing """

    # TODO: mark content as viewed by this wb when sent full list
    # TODO: keep track of sessions / cookies
    # TODO: implement user preferences

    # Don't get massive amounts of random posts
    if count > 30:
        return {'status':'error','code':500,'data':'Requested too many posts'}

    all_keys = get_keys()
    if len(all_keys) < count:
        return {'status':'error','code':500,'data':'Basically empty datastore'}

    posts = get_posts(all_keys, count)
    images = get_post_images(posts)
    data = format_response(posts, images)
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
