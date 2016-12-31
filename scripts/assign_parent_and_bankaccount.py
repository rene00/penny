"""
Link parents with orphans based on orphan account and parent memo.

On 2016-05-21 I discovered that the peewee to SQLA migration had not
set the parents or bank accounts for child transactions, leaving
the children as orphans.

This failure to link children with parents resulted in incorrect
bank account balances as orphans did not have a bank account assigned
yet the parents which had their amounts adjusted to take into consideration
their children transactions had no children.
"""

import sys
import os
sys.path.insert(1, os.path.join(sys.path[0], '..'))

from app import app, models
from app.models import session
from flask.ext.script import Manager
from sqlalchemy.orm.exc import NoResultFound

manager = Manager(app, usage='Import Myer Visa Transactions.')


def _suggest_link(parent, orphan):
    print("parent: {0.id}, {0.date}, {0.memo}".format(parent))
    print("orphan: {0.id}, {0.date}, {0.memo}".format(orphan))
    msg = "Link parent and orphan?"
    shall = raw_input("%s " % msg).lower() == 'y'
    if shall:
        orphan.parent_id = parent.id
        orphan.bankaccount = parent.bankaccount
        session.commit()


@manager.option('--child-account-id', dest='child_account_id', required=True)
@manager.option('--parent-memo', dest='parent_memo', required=True)
def process_transactions(child_account_id, parent_memo):

    # Select all parent transactions.
    parents = session.query(models.Transaction). \
        filter_by(memo=parent_memo).order_by(models.Transaction.date.asc()).all()

    # Select account.
    try:
        account = session.query(models.Account). \
            filter_by(id=child_account_id).one()
    except NoResultFound:
        app.logger.error('Unable to find child account')
        sys.exit(2)

    # Loop through parents, find the orphans and suggest the link.
    for parent in parents:
        orphans = session.query(models.Transaction). \
            filter_by(date=parent.date, bankaccount=None, account=account). \
            all()
        for orphan in orphans:
            _suggest_link(parent, orphan)

if __name__ == '__main__':
    manager.run()
