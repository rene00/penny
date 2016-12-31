import sys
import os
sys.path.insert(1, os.path.join(sys.path[0], '..'))

from app import app
from app.models import session, Transaction, BankAccount
import hashlib
from flask.ext.script import Manager
from sqlalchemy.orm.exc import NoResultFound


manager = Manager(app)

def get_hash(transaction):
    _hash = hashlib.md5()
    for key in [transaction.date, transaction.debit, transaction.credit, transaction.memo]:
        if key:
            _hash.update(str(key))
    return _hash.hexdigest()


@manager.option('--bankaccount-id', dest='bankaccount_id', default=15)
def run(bankaccount_id):

    try:
        bankaccount = session.query(BankAccount). \
            filter_by(id=bankaccount_id).one()
    except NoResultFound:
        raise

    transactions = session.query(Transaction).filter(
            Transaction.bankaccount == bankaccount,
            Transaction.parent_id == None,
            Transaction.is_deleted == 0).all()
    track = {}

    for transaction in transactions:
        _hash = get_hash(transaction)
        if _hash not in track:
            track[_hash] = transaction.id
        else:
            t = session.query(Transaction).filter(Transaction.id == track[_hash]).one()
            print("dupe: {0}, {1}".format(t.id, transaction.id))
            print(t)
            print(transaction)
            msg = 'Delete {0}?'.format(transaction.id)
            shall = raw_input("%s (y/N) " % msg).lower() == 'y'
            if shall:
                transaction.is_deleted = True
                session.commit()

if __name__ == '__main__':
    manager.run()
