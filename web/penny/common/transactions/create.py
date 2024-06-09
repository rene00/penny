from flask import current_app
from nene import db
from nene.transaction import get_tx_hash
from nene.memc import Memcache


class CreateTransactionError(StandardError):
    pass


class CreateTransaction:
    def __init__(
        self,
        date,
        memo,
        bankaccount,
        user,
        account=None,
        fitid=None,
        debit=0,
        credit=0,
        paypalid=None,
        parent=None,
    ):
        self.date = date
        self.debit = debit
        self.credit = credit
        self.memo = memo
        self.bankaccount = bankaccount
        self.user = user
        self.account = account
        self.fitid = fitid
        self.paypalid = paypalid
        self.parent = parent

        # Check if bankaccount is a DB object and if so pass in the id
        # attribute to get_tx_hash(). If bankaccount is not a DB object,
        # dont pass it through to get_tx_hash() as it defaults to False.
        if hasattr(self.bankaccount, "id"):
            self.tx_hash = get_tx_hash(
                self.date,
                self.debit,
                self.credit,
                self.memo,
                self.fitid,
                self.paypalid,
                self.bankaccount.id,
            )
        else:
            self.tx_hash = get_tx_hash(
                self.date, self.debit, self.credit, self.memo, self.fitid, self.paypalid
            )

    def create(self):
        # confirm bankaccount is owned by the user.
        if self.bankaccount:
            if self.bankaccount.user != self.user:
                raise CreateTransaction(
                    "user does not own bankaccount; user={}, bankaccount={}".format(
                        self.user.id, self.bankaccount.id
                    )
                )

        # account is optional though if submitted check that the account
        # exists and it is owned by the user.
        if self.account and self.account.user != self.user:
            raise CreateTransactionError(
                "user does not own account or account doest not exist; "
                "user={}, account={}".format(self.user.id, self.account.id)
            )

        self.transaction = None
        try:
            self.transaction = db.Transaction.get(
                db.Transaction.hash == self.tx_hash, db.Transaction.user == self.user
            )
        except db.Transaction.DoesNotExist:
            self.transaction = db.Transaction(
                date=self.date,
                debit=self.debit,
                credit=self.credit,
                memo=self.memo,
                bankaccount=self.bankaccount,
                account=self.account,
                hash=self.tx_hash,
                fitid=self.fitid,
                paypalid=self.paypalid,
                user=self.user,
                parent=self.parent,
            )
            self.transaction.save()
        else:
            # A transaction which shares the same hash exists. If the
            # account is different update it else consider it a
            # duplicate. Log this for now though in the future something
            # should be done with duplicates such as creating a report
            # or providing the opportunity to import it.
            # trac@48: Handle duplicate transactions.
            if self.account and self.account != self.transaction.account:
                self.transaction.account = self.account.id
                self.transaction.save()
                mc = Memcache()
                mc.delete("transaction.{}".format(self.transaction.id))
            else:
                # I'm not convinced this is being reached and logged.
                # XXX: investigate if this is being logged.
                current_app.logger.info(
                    "duplicate transaction found on import; "
                    "transaction={transaction}, user={user}".format(
                        transaction.self.transaction.id, user=self.user.id
                    )
                )
        finally:
            return self.transaction
