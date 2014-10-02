import api as _api
import bottle
import cron as _cron

# http://bottlepy.org/docs/dev/tutorial.html

# Have bottle run in debug mode (better error messages)
bottle.debug(True)


# API middleware
# https://github.com/smthmlk/restware
# TODO: pretty print: https://pypi.python.org/pypi/bottle-api-json-formatting/0.1.1


# API call handlers
api = bottle.Bottle(autojson=True)

@api.get('/api/v1/post/random')
@api.get('/api/v1/post/random/')
@api.get('/api/v1/post/random/<count:int>')
def api_post_random(count=10):
    return _api.post_random(count)

@api.get('/api/v1/img/<urlsafe>')
def api_image_info(urlsafe=None):
    return _api.image_info(urlsafe)

@api.get('/api/v1/tags')
@api.get('/api/v1/tags/')
def api_tags():
    return _api.tags()

@api.error(404)
def api_error(err):
    print str(err)
    new_err = dict(status='error', code=404, data='Endpoint not found')
    # return new_err

    # PATCH: Bottle doen't return json correctly
    from json import dumps
    bottle.response.content_type = 'application/json'
    return dumps(new_err)


# CRON call handlers
cron = bottle.Bottle()

@cron.get('/cron/parse_feeds')
def cron_parse_feeds():
    return _cron.parse_feeds()


# ERR call handlers
err = bottle.Bottle()

@err.error(404)
@cron.error(404)
def err_error(err):
    print str(err)
    return 'Your kung-fu is not strong'
