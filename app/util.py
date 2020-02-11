import logging
import sys
import locale


def get_logger_handle():
    """Return a logging handler."""
    handle = logging.StreamHandler(stream=sys.stdout)
    formatter = logging.Formatter(
        "%(asctime)s - %(pathname)s:%(lineno)d - %(levelname)s - %(message)s")
    handle.setFormatter(formatter)
    handle.setLevel(logging.INFO)
    return handle


def convert_to_float(cents):
    locale.setlocale(locale.LC_ALL, 'en_AU.UTF-8')
    return locale.currency(float(cents / float(100)), grouping=True)
