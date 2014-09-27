import chive
from app.models import Post

import bottle

bottle.DEBUG = True  # Monkey-patching

cron = bottle.Bottle()

# TODO: Check if posts are updated spairingly...

@cron.get('/cron/parse_feeds')
def parse_feeds():
    """ parse rss feeds into the datastore """
    page_count = 1
    hit_count = 0

    # TODO: track individual images

    # Loop over all the feeds
    for feed in chive.next_page():
        for item in feed.items:

            # Generate ID for feeds
            item_id = item.guid

            # Check for item existing in datastore (insert if necessary)
            if Post.get_by_id(item_id):
                hit_count += 1
            else:
                post_data = item.to_dict()
                post = Post(id=item_id, **post_data)
                post.put()

        # Run into 5 already found feeds and break
        if hit_count >= 5:
            break

        print page_count
        page_count += 1

        # Development
        if page_count > 10:
            break
    return 'done'
