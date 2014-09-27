from app.models import Post, Img
import random


def data():
    """ process new chive articles """

    # TODO: mark content as viewed by this wb when sent full list
    # TODO: keep track of sessions / cookies
    # TODO: implement user preferences
    # TODO: pretty print: https://pypi.python.org/pypi/bottle-api-json-formatting/0.1.1

    # Thanks: http://stackoverflow.com/a/21650400
    query = Post.query()
    all_keys = query.fetch(keys_only=True)
    if len(all_keys) < 10:
        return {'status':'error', 'data': 'Basically empty datastore'}

    # Get 10 random posts
    list_keys = random.sample(all_keys, 10)
    posts = [key.get() for key in list_keys]

    # Populate posts with images
    for post in posts:
        images = []
        for key in post.keys:
            img = key.get().to_dict()
            images.append(img)
        post.media = images

    # Convert objects to dicts
    exclude = ['keys']
    data = [post.to_dict(exclude=exclude) for post in posts]

    return {'status':'success','code':200,'data':data}


def meta(img=None):
    """ Update image metadata (server never loads images) """
    return 'TODO %s' % img
