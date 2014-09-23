import chive
from app.models import Post


def main():
    page_count = 1
    hit_count = 0

    for feed in chive.next_page():
        for item in feed.items:

            if Post.get_by_id(id):
                hit_count += 1
            else:
                post_data = item.to_dict()
                post = Post(id=item.guid, **post_data)
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
