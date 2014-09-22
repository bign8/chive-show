from lxml import etree
from urllib2 import urlopen
import json
import re


title_dirty = re.compile(r'^[a-zA-Z0-9-]*$')


class RSS:
    def __init__(self, url):
        """ Initialize rss feed """
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
            self.items.append(_rss_item(item))

        # Final attributes
        self.has_items = len(self.items) > 0

    def to_dict(self):
        """ Convert this bitch to a dictionary """
        dic = self.__dict__
        dic['items'] = [x.to_dict() for x in self.items]
        return dic


class _rss_item:
    def __init__(self, item):
        """ Process feed item """
        ns = {'namespaces': item.nsmap}

        # Meta
        self.title = item.xpath('title/text()')[0]
        self.link = item.xpath('link/text()')[0]
        self.published = item.xpath('pubDate/text()')[0]
        self.categories = item.xpath('category/text()')
        self.guid = item.xpath('guid/text()')[0]

        # Creator
        self.creator = {
            'name': item.xpath('dc:creator/text()', **ns)[0],
            'img': None,
        }

        # Content (append to creator)
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
    def __init__(self, content):
        ns = {'namespaces': content.nsmap}

        # URL
        self.url = content.get('url')

        # Category
        category = content.xpath('media:category/text()', **ns)
        self.category = category[0] if category else None

        # Rating
        rating = content.xpath('media:rating/text()', **ns)
        self.rating = rating[0] if rating else None

        # Title
        title = content.xpath('media:title/text()', **ns)
        # title = None if title_dirty.match(title) else title
        # TODO: process title here (don't gather dumb ones)
        self.title = title[0] if title else None

        if self.title and title_dirty.match(self.title):
            self.title = None

    def to_dict(self):
        return self.__dict__
