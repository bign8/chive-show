# from cron import main
from models import Post
import bottle
import random

bottle.DEBUG = True

app = bottle.Bottle()


# Static Router
@app.route('/static/<filepath:path>')
def server_static(filepath):
    """ Static file server """
    return bottle.static_file(filepath, root='./static')


@app.route('/')
def index():
    return bottle.static_file('index.html', root='./static')


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
    list_keys = random.sample(all_keys, 10)
    return {'items': [key.get().to_dict() for key in list_keys]}


# For development only
# @app.get('/cron')
# def cron():
#     main()


@app.get('/api/meta/<img>')
def update_meta(img=None):
    """ Update image metadata (server never loads images) """
    return 'TODO %s' % img


@app.error(404)
def error(err):
    """ Display error message / hacker message """
    print str(err)
    return 'Your kung-fu is not strong'


# NON-DB Development: Local develoment server (without GAE dev server)
if __name__ == "__main__":
    app.run(host='localhost', port=8080, debug=True)
