__author__ = 'bign8'

import bottle

# Monkey-patching bottle to run in debug mode (better error messages)
bottle.debug(True)


# API middleware
# https://github.com/smthmlk/restware


# API call handlers
api = bottle.Bottle(autojson=True)

@api.get('/api/data')
def api_data():
    from main import data
    return data()

@api.get('/api/meta/<img>')
def api_meta(img=None):
    from main import meta
    return meta(img)

@api.error(404)
def api_error(err):
    print str(err)
    new_err = dict(status='error', code=404, data='Endpoint not found')

    # PATCH: Bottle doen't return json correctly
    from json import dumps
    bottle.response.content_type = 'application/json'
    return dumps(new_err)


# CRON call handlers
cron = bottle.Bottle()

@cron.get('/cron/parse_feeds')
def cron_parse_feeds():
    from cron import parse_feeds
    return parse_feeds()


# ERR call handlers
err = bottle.Bottle()

@err.error(404)
@cron.error(404)
def err_error(err):
    print str(err)
    return 'Your kung-fu is not strong'
