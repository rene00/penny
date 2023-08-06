import time
import hashlib
import datetime
import re
from penny import models


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


def generate_transaction_hash(t: models.Transaction) -> str:
    """Generate a SHA256 hash of the transaction."""

    memo: str = re.sub(r"\s\s+", " ", t.memo)
    memo = re.sub(r"\s$", "", memo)

    _hash = hashlib.sha256()
    for param in (t.date, t.debit, t.credit, memo, t.fitid, t.bankaccount.id):
        # If param is datetime convert to str.
        if isinstance(param, datetime.datetime):
            param = param.isoformat()
        # If param is an int convert to str.
        if isinstance(param, int):
            param = str(param)
        # If param is None, skip it.
        if param is None:
            continue
        _hash.update(param.encode("utf-8"))
    return _hash.hexdigest()
