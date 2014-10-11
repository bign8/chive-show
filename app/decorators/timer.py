import logging
from time import time

from .decorator import Decorator


class Timer(Decorator):
    def setup(self, name, level='debug'):
        self.name =  name
        self.level = level

    def before(self):
        self.start = time()

    def after(self, result):
        # TODO: set logging level
        delta = time() - self.start
        count = len(result)
        log = logging.debug
        log(
            '%s took %fs and returned %i values (%f items/s)',
            self.name, delta, count, count / delta
        )
