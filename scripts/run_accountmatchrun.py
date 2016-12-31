import sys
import os
sys.path.insert(1, os.path.join(sys.path[0], '..'))

from flask.ext.script import Manager
from app import app, models
from app.common.accountmatchrun import AccountMatchRun

manager = Manager(app)


@manager.option('--user-id', dest='user_id')
def run(user_id):
    user = models.User.query.get(user_id)
    _accountmatchrun = AccountMatchRun(user)
    print(_accountmatchrun.run())

if __name__ == '__main__':
    manager.run()
