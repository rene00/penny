import logging
import sys


def get_logger_handle():
    """Return a logging handler."""
    handle = logging.StreamHandler(stream=sys.stdout)
    formatter = logging.Formatter(
        "%(asctime)s - %(pathname)s:%(lineno)d - %(levelname)s - %(message)s")
    handle.setFormatter(formatter)
    handle.setLevel(logging.INFO)
    return handle
