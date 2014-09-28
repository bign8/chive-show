from rss import RSS


# TODO: Parse all of the chives backlog (archive) and store in plain text
#    in case rss parsing changes

def _cleanup(feed):
    # TODO: clean image count from titles of posts
    return feed


def next_page():
    """ Generate the next chive feed """

    # Generate first feed
    FIRST = 'http://thechive.com/feed/?page'
    feed = RSS(FIRST)
    yield _cleanup(feed)

    # Generate next feed
    index = 2
    MORE = 'http://thechive.com/feed/?paged=%s'
    while feed.has_items:
        feed = RSS(MORE % index)
        index += 1
        yield _cleanup(feed)