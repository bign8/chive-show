import chive
# from history import get_history
from app.models import Post, Img
from google.appengine.ext import ndb

# TODO: Check if posts are updated spairingly...
# TODO: deal with instance timeouts
#       parse new posts until hit_count >= 5
#       start processing history at last known search page (ignore hit_count)
#       do all this while monitoring the time <= 60seconds

def parse_feeds(start=0):
    """ parse rss feeds into the datastore """
    hit_count = 0
    page_count = 0

    # Loop over all the feeds
    for feed in chive.next_page(start):

        page_count += 1
        hit_count += _process_items(feed)

        # Kill if 5 collisions or 5 pages
        if hit_count >= 5 or page_count >= 5:
            break

    # TODO: check this works
    # get_history()

    return 'done'

def _process_feed_posts(feed):
    # Check if Posts exist (feed) => posts
    post_keys = [ndb.Key(Post, post.guid) for post in feed.items]
    posts_dbo = ndb.get_multi(post_keys)  # DB Call (1)
    posts_dirty = zip(post_keys, feed.items, posts_dbo)
    posts_mod = [Post(key=post[0], **post[1].to_dict()) for post in posts_dirty]

    posts_dirty = zip(post_keys, feed.items, posts_dbo, posts_mod)
    posts = [post for post in posts_dirty if not post[2]]
    return posts, len(posts_dirty) - len(posts)

def _process_posts_media(posts):
    # Check if Images exist (posts) => images
    images_multi = [item[1].media for item in posts]
    images_rss = _flatten(images_multi)
    images_key = [ndb.Key(Img, img['url']) for img in images_rss]
    images_db = ndb.get_multi(images_key)  # DB Call (2)

    images = zip(images_key, images_rss)
    images_mod = [Img(key=img[0], **img[1]) for img in images]
    images_tostore = [not img for img in images_db]

    # Key, Model, To-Store
    return zip(images_key, images_mod, images_tostore)

def _store_posts_media(images):
    # Insert images into DB (images) => None
    image_insert = [img[1] for img in images if img[2]]  # for putting
    ndb.put_multi(image_insert)  # DB Call (3)

def _store_feed_posts(posts, images):
    # maybe TODO: have posts class contain a number of keys (along with _rss_item)
    # Populate posts with images (posts, images) => None
    start = 0
    post_models = [post[3] for post in posts]
    for idx, post in enumerate(post_models):
        end = start + len(posts[idx][1].media)
        post.keys = [img[0] for img in images[start:end]]
        start = end

    ndb.put_multi(post_models)  # DB Call (4)

def _flatten(data):
    return [element for sublist in data for element in sublist]

def _process_items(feed):
    # TODO: break logic up into above functions
    # TODO: pass between functions (key, rss_model, db_model, bool (to store in db))

    # Index orders
    # 0. Key - database key instance
    # 1. RSS - RSS object associated with db element
    # 2. DBo - Data pulled from db for that key
    # 3. MOD - TODO: Model to be put in db

    posts, hit_count = _process_feed_posts(feed)

    if not posts:
        return 99

    images = _process_posts_media(posts)

    _store_posts_media(images)

    _store_feed_posts(posts, images)

    return hit_count
