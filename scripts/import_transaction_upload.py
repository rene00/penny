import sys
import os
sys.path.insert(1, os.path.join(sys.path[0], '..'))

from app import app
from app.common.tasks import import_transactions
from flask.ext.script import Manager

manager = Manager(app)


@manager.option('--transactionupload-id', dest='transactionupload_id')
@manager.option('--user-id', dest='user_id')
def do_upload(transactionupload_id, user_id):
    print(import_transactions(transactionupload_id, int(user_id)))


if __name__ == '__main__':
    manager.run()
