from app.models import session, User, TransactionUpload, BankAccount
from app.common import filetypes
from app.common.import_transactions import (ImportTransactions,
                                            ImportTransactionsError)
from app.common.accountmatchrun import AccountMatchRun
from flask_rq import job
from sqlalchemy.orm.exc import NoResultFound


@job
def import_transactions(id, user_id):
    from app import app as _app
    with _app.app_context():
        filetype = None

        try:
            transactionupload = session.query(TransactionUpload). \
                filter_by(id=id, user_id=user_id).one()
        except NoResultFound:
            raise

        try:
            user = session.query(User).filter_by(id=user_id).one()
        except NoResultFound:
            raise

        try:
            if filetypes.is_ofx(transactionupload.filepath):
                filetype = 'ofx'
        except filetypes.FileTypesError:
            raise

        bankaccount_number = filetypes.get_bankaccount_number_from_ofx(
            transactionupload.filepath)

        try:
            bankaccount = session.query(BankAccount) \
                .filter_by(number=bankaccount_number, user=user).one()
        except NoResultFound:
            bankaccount = BankAccount(user_id=user_id,
                                      number=bankaccount_number)
            session.add(bankaccount)
            session.commit()

        # Read all transactions into a list
        transactions = None
        with open(transactionupload.filepath, 'r') as fh:
            transactions = fh.read()

        # Instantiate the ImportTransactions object ready to import.
        import_transactions = ImportTransactions(transactions, bankaccount,
                                                 user)

        # Import the transactions.
        imported_transactions = None
        if filetype == 'ofx':
            try:
                imported_transactions = import_transactions.process_ofx()
            except ImportTransactionsError:
                raise

        return imported_transactions


@job
def run_accountmatchrun(user_id):
    from app import app as _app
    with _app.app_context():
        try:
            user = session.query(User).filter_by(id=user_id).one()
        except NoResultFound:
            raise
        else:
            _accountmatchrun = AccountMatchRun(user)
            return _accountmatchrun.run()
