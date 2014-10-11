from lxml import etree
from urllib2 import urlopen
import json
import re
# import datetime
# import time


class RSS:
    def __init__(self, url, deep=True):
        """
        Initialize rss feed
        :param deep: To store everything or just what's needed for searching
        """
        stream = urlopen(url)
        tree = etree.parse(stream, etree.XMLParser(recover=True))
        root = tree.getroot()

        # Feed Meta
        self.title = root.xpath('//channel/title/text()')[0]
        self.description = root.xpath('//channel/description/text()')[0]
        self.image = root.xpath('//channel/image/url/text()')[0]
        self.link = root.xpath('//channel/link/text()')[0]

        # Process Feed Items
        self.items = []
        for item in root.xpath('//channel/item'):
            self.items.append(_rss_item(item, deep))

        # Final attributes
        self.has_items = len(self.items) > 0

    def to_dict(self):
        """ Convert this bitch to a dictionary """
        dic = self.__dict__
        dic['items'] = [x.to_dict() for x in self.items]
        return dic


class _rss_item:
    def __init__(self, item, deep=True):
        """
        Process feed item
        :param deep: To store everything or just what's needed for searching
        """
        ns = {'namespaces': item.nsmap}

        # Meta
        self.title = item.xpath('title/text()')[0]
        self.link = item.xpath('link/text()')[0]
        self.tags = item.xpath('category/text()')
        self.guid = item.xpath('guid/text()')[0]

        # Date
        self.date = item.xpath('pubDate/text()')[0] # Tue, 23 Sep 2014 01:00:01 +0000
        # struct = time.strptime(date, '%a, %d %b %Y %H:%M:%S +0000')
        # self.date = datetime.datetime(*struct, 0)

        # Creator
        self.creator = {
            'name': item.xpath('dc:creator/text()', **ns)[0],
            'img': None,
        }

        # Content (append to creator)
        if deep:
            self.media = []
            for content in item.xpath("media:content[@medium='image']", **ns):
                media = _rss_media(content)
                if media.category and 'author' in media.category:
                    self.creator['img'] = media.url
                else:
                    self.media.append(media)

    def to_dict(self):
        dic = self.__dict__
        dic['media'] = [x.to_dict() for x in self.media]
        return dic


class _rss_media:
    dirty_title = re.compile(r'^[a-zA-Z0-9-]*$')

    def __init__(self, content):
        ns = {'namespaces': content.nsmap}

        # URL
        self.url = content.get('url')

        # Category
        category = content.xpath('media:category/text()', **ns)
        self.category = category[0] if category else []

        # Rating
        rating = content.xpath('media:rating/text()', **ns)
        self.rating = rating[0] if rating else None

        # Title
        title = content.xpath('media:title/text()', **ns)
        self.title = title[0] if title else None

        if self.title and self.dirty_title.match(self.title):
            self.title = None

    def to_dict(self):
        return self.__dict__
