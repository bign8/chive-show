from rss import RSS

FIRST_PAGE = 'http://thechive.com/feed/?page'
PAGE_TEMPLATE = 'http://thechive.com/feed/?paged={}'

# TODO: Parse all of the chives backlog (archive) and store in plain text
#    in case rss parsing changes

def page_url(idx):
    """ Return page url for a specific feed index """
    return PAGE_TEMPLATE.format(idx - 1) if idx > 0 else FIRST_PAGE


def cleanup(feed):
    # TODO: clean image count from titles of posts
    return feed


def page_rss(idx, deep=True):
    """ Return page rss for a specific feed index """
    url = page_url(idx)
    feed = RSS(url, deep)
    return cleanup(feed)


def next_page(start=0):
    """ Generate the next chive feed """

    for idx in xrange(start, 3 ** 10):
        yield page_rss(idx)
