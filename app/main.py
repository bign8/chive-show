from models import Post
import bottle
import random

bottle.DEBUG = True  # Monkey-patching!!!

app = bottle.Bottle()

# API middleware
# https://github.com/smthmlk/restware


@app.get('/api/data')
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
    list_keys = random.sample(all_keys, 10)
    return {'status':'success','data': [key.get().to_dict() for key in list_keys]}


@app.get('/api/meta/<img>')
def update_meta(img=None):
    """ Update image metadata (server never loads images) """
    return 'TODO %s' % img


@app.error(404)
def error(err):
    """ Display error message / hacker message """
    print str(err)
    return 'Your kung-fu is not strong'
