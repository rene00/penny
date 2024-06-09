import sys
import os

sys.path.insert(1, os.path.join(sys.path[0], ".."))

from app import app
from app.models import session, Transaction, BankAccount
import hashlib
from flask.ext.script import Manager
import re


manager = Manager(app)


@manager.option("--bankaccount-id", dest="bankaccount_id", default=15)
def run(bankaccount_id):

    try:
        bankaccount = session.query(BankAccount).filter_by(id=bankaccount_id).one()
    except NoResultFound:
        raise

    transactions = (
        session.query(Transaction)
        .filter(
            Transaction.bankaccount == bankaccount,
            Transaction.parent_id == None,
            Transaction.is_deleted == 0,
        )
        .all()
    )
    track = {}

    for transaction in transactions:

        memo = transaction.memo
        memo = re.sub("\s\s+", " ", memo)
        memo = re.sub("\s$", "", memo)
        print(memo)

        if memo != transaction.memo:
            transaction.memo = memo
            session.add(transaction)
            session.commit()
            transaction.transactionm_hash = transaction.get_hash(transaction)
            session.add(transaction)
            session.commit()
            print('Updated "{0}" to "{1}"'.format(transaction.memo, memo))


if __name__ == "__main__":
    manager.run()
