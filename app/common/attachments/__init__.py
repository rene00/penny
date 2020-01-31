import os
from werkzeug import secure_filename
import hashlib


def mkdir_for_transaction(parentdir, relfilepath):
    filedir = os.path.join(parentdir, os.path.dirname(relfilepath))
    if not os.path.isdir(filedir):
        os.makedirs(filedir)


def get_filepath_for_transaction_attachment(parentdir, transaction, filename):
    "Return absolute and relative filepath of transaction attachment."
    relfilepath = os.path.join(
        str(transaction.id),
        secure_filename(filename)
    )
    mkdir_for_transaction(parentdir, relfilepath)
    return (
        os.path.join(parentdir, relfilepath),
        os.path.join('/', relfilepath)
    )


def get_hash_of_file(filepath):
    return hashlib.md5(open(filepath, 'rb').read()).hexdigest()
