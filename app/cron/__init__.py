import chive
from app.models import Post, Img

# TODO: Check if posts are updated spairingly...
# TODO: deal with instance timeouts
#       parse new posts until hit_count >= 5
#       start processing history at last known search page (ignore hit_count)
#       do all this while monitoring the time <= 60seconds

def parse_feeds():
    """ parse rss feeds into the datastore """
    hit_count = 0
    page_count = 0

    # Loop over all the feeds
    for feed in chive.next_page():

        page_count += 1
        print page_count
        hit_count += _process_items(feed)

        # Kill if 5 collisions or 5 pages
        if hit_count >= 5 or page_count >= 5:
            break

    return 0


def _process_items(feed):
    hit_count = 0
    for item in feed.items:
        # Generate ID for feeds
        item_id = item.guid

        # Check for item existing in datastore (insert if necessary)
        if Post.get_by_id(item_id):
            hit_count += 1
        else:

            # Store images
            img_keys = _store_images(item)

            # Store Posts
            post_data = item.to_dict()
            post = Post(id=item_id, **post_data)
            post.keys = img_keys
            post.put()

    return hit_count


def _store_images(item):
    img_keys = []

    for img in item.media:
        img_id = img.url

        ele = Img.get_by_id(img_id)

        if ele:
            img_key = ele.key
        else:
            img_data = img.to_dict()
            img = Img(id=img_id, **img_data)
            img_key = img.put()

        img_keys.append(img_key)

    return img_keys
