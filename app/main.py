from bottle import Bottle, static_file, redirect
import chive

app = Bottle()


# Static Router
@app.route('/static/<filepath:path>')
def server_static(filepath):
    """ Static file server """
    return static_file(filepath, root='./static')


@app.route('/')
def index():
    return static_file('index.html', root='./static')
    # TODO: remove this line and include (for reference)
    redirect('/static/index.html')


@app.get('/api/data')
def cron():
    """ process new chive articles """

    # TODO: Store in DB
    # TODO: mark content as viewed by this wb when sent full list
    # TODO: keep track of sessions / cookies
    # TODO: implement user preferences

    # More tuned for cron, just a test
    for feed in chive.next_page():
        return feed.to_dict()


@app.get('/api/meta/<img>')
def update_meta(img=None):
    """ Update image metadata (server never loads images) """
    return 'TODO %s' % img


@app.error(404)
def error(err):
    """ Display error message / hacker message """
    print str(err)
    return 'Your kung-fu is not strong'


# Local develoment server (without GAE dev server)
if __name__ == "__main__":
    app.run(host='localhost', port=8080, debug=True)
