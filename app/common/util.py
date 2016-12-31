import time


def now(delta=None):
    "return now as unixtime in milliseconds."
    if delta is None:
        return int((time.time() + 0.5) * 1000)
    else:
        now = int((time.time() + 0.5) * 1000)
        return now + delta


def merge_dicts(*dict_args):
    """Given two dicts, merge them into a new dict as a shallow copy."""
    result = {}
    for dictionary in dict_args:
        result.update(dictionary)
    return result
