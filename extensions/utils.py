import fcntl
import functools


class Lock(object):
    def __init__(self, path):
        self._path = path

    def __enter__(self):
        self._fd = open(self._path, 'w')
        fcntl.flock(self._fd, fcntl.LOCK_EX)

    def __exit__(self, type, value, traceback):
        fcntl.flock(self._fd, fcntl.LOCK_UN)
        self._fd.close()


class exclusive(object):  # NOQA
    def __init__(self, path):
        self._path = path

    def __call__(self, fnc):
        @functools.wraps(fnc)
        def wrapper(*args, **kwargs):
            with Lock(self._path):
                return fnc(*args, **kwargs)

        return wrapper
