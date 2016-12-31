import os
from werkzeug import secure_filename
import hashlib


def mkdir_for_transaction(filepath):
    filedir = os.path.dirname(os.path.realpath(filepath))
    if not os.path.isdir(filedir):
        os.makedirs(filedir)


def get_filepath_for_transaction_attachment(app, transaction, filename):
    filepath = os.path.join(
        app.config['TRANSACTION_ATTACHMENTS_UPLOAD_FOLDER'],
        str(transaction.id),
        secure_filename(filename))
    mkdir_for_transaction(filepath)
    return filepath


def get_hash_of_file(filepath):
    return hashlib.md5(open(filepath, 'rb').read()).hexdigest()
