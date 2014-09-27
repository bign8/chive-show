import chive
from app.models import Post, Img

# TODO: Check if posts are updated spairingly...

def parse_feeds():
    """ parse rss feeds into the datastore """
    page_count = 1
    hit_count = 0

    # Loop over all the feeds
    for feed in chive.next_page():
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

        # Run into 5 already found feeds and break
        if hit_count >= 5:
            break

        print page_count
        page_count += 1

        # Development
        if page_count > 10:
            break
    return 'done'


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
