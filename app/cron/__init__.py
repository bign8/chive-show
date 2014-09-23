import chive
from app.models import Post

# TODO: move whatever calls this off to models
from google.appengine.ext import ndb


def main():
    page_count = 1
    hit_count = 0

    for feed in chive.next_page():
        for item in feed.items:

            # TODO: check for updated posts once a day
            item_key = ndb.Key('Post', item.guid)
            item_obj = Post.get_by_id(item_key.id())

            if item_obj:
                hit_count += 1

            post_data = item.to_dict()
            post = Post(key=item_key, **post_data)
            post.put()

        # Run into 5 already found feeds and break
        if hit_count >= 5:
            break

        print page_count
        page_count += 1

        # Development
        if page_count >= 10:
            break
    return 'done'
