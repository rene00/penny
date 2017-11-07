from io import StringIO
from app import models
from app.common import currency
from bs4 import BeautifulSoup
from ofxparse import OfxParser
from ofxparse.ofxparse import OfxParserException
from sqlalchemy.exc import IntegrityError
import logging as log
import re


class ImportTransactionsError(Exception):
    pass


class ImportTransactions():
    def __init__(self, transactions, bankaccount, user):
        self.transactions = transactions
        self.bankaccount = bankaccount
        self.user = user

    def process_ofx(self):
        # List of transactions that have been processed and will be
        # returned.
        transactions = []

        # roll ofx file through beautifulsoup first to correct any
        # issues.
        soup = BeautifulSoup(self.transactions, 'html.parser')
        try:
            ofx = OfxParser.parse(StringIO(soup))
        except OfxParserException:
            raise
        except UnicodeDecodeError:
            raise ImportTransactionsError('failed to import file.')

        # Check that bankaccount numbers match. The number extracted
        # from the ofx file are the ofc routing and account numbers
        # joined.
        if ofx.account.routing_number and ofx.account.number:
            account_number = '{}{}'.format(ofx.account.routing_number,
                                           ofx.account.number)
            if str(account_number).lower() != \
                    str(self.bankaccount.number).lower():
                raise ImportTransactionsError(
                    'bankaccount numbers dont match; '
                    'ofx={}, bankaccount={}'.format(account_number,
                                                    self.bankaccount.number)
                )

        for tx in ofx.account.statement.transactions:
            transaction = models.Transaction(date=tx.date, memo=tx.memo,
                                             bankaccount=self.bankaccount,
                                             fitid=None, user=self.user)

            # Strip double whitespace from transaction memo.
            transaction.memo = re.sub("\s\s+", " ", transaction.memo)

            # Strip leading whitespace from transaction memo.
            transaction.memo = re.sub("\s$", "", transaction.memo)

            # If the OFX id exists, set is as the fitid for the
            # transaction.
            if tx.id:
                transaction.fitid = tx.id

            # Convert amount to cents.
            amount = currency.to_cents(tx.amount)

            # Set credit and debit for the transaction.
            (transaction.credit, transaction.debit) = \
                currency.get_credit_debit(amount)

            # Set the transaction hash.
            transaction.transaction_hash = transaction.get_hash(transaction)

            models.db.session.add(transaction)

            log.info("About to commit transaction: transaction={0}"
                     .format(transaction))

            try:
                models.db.session.commit()
            except IntegrityError:
                # Ignore duplicate transactions.
                models.db.session.rollback()
            else:
                # Append new transaction to list.
                transactions.append(transaction)

        return transactions
