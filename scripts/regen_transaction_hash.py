import sys
import os

sys.path.insert(1, os.path.join(sys.path[0], ".."))

from app import app
from app.models import session, Transaction, BankAccount
from app.common import util
import hashlib
from flask.ext.script import Manager
import re
import sqlalchemy
import time


manager = Manager(app)


@manager.command
def run():

    transactions = (
        session.query(Transaction)
        .filter(
            Transaction.parent_id == None,
            Transaction.is_deleted == 0,
        )
        .all()
    )
    track = {}

    for transaction in transactions:

        print(transaction.id)
        new_hash = util.generate_transaction_hash(
            date=transaction.date,
            debit=transaction.debit,
            credit=transaction.credit,
            memo=transaction.memo,
            fitid=transaction.fitid,
            bankaccount_id=transaction.bankaccount.id,
        )

        if new_hash != transaction.transaction_hash:
            print(transaction.transaction_hash, new_hash)
            transaction.transaction_hash = new_hash
            session.add(transaction)
            try:
                session.commit()
            except sqlalchemy.exc.IntegrityError:
                # transaction.is_deleted = 1
                # session.add(transaction)
                session.rollback()
                session.commit()
                print("Collision {0}".format(transaction.id))

                # Find other transaction which shares hash
                dupe = (
                    session.query(Transaction)
                    .filter(Transaction.transaction_hash == new_hash)
                    .one()
                )
                session.delete(dupe)
                session.commit()
                transaction.transaction_hash = new_hash
                session.add(transaction)
                session.commit()
            else:
                print(
                    'Updated {2} "{0}" to "{1}"'.format(
                        transaction.transaction_hash, new_hash, transaction.id
                    )
                )
            # time.sleep(1)


if __name__ == "__main__":
    manager.run()
