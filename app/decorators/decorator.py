

class Decorator:
    def __init__(self, *args, **kwargs):
        if self.setup:
            self.setup(*args, **kwargs)

    def __call__(self, func):
        self.func = func
        return self.action

    def action(self, *args, **kwargs):
        if self.before:
            self.before()

        result = self.func(*args, **kwargs)

        if self.after:
            self.after(result)

        return result
