from flask import current_app as app
from app import models
from app.common.util import merge_dicts


class ReportsProfitLoss():

    def __init__(self, **kwargs):
        self.entity = kwargs.get('entity')
        self.bankaccount = kwargs.get('bankaccount')
        self.title = 'Profit Loss'
        self.start_date = kwargs.get('start_date')
        self.end_date = kwargs.get('end_date')
        self.order_by = kwargs.get('order_by', 'account')

    def generate(self):

        txs = None

        self.transactions = ProfitLossTransactions()

        app.logger.info(
            'Generating report; bankaccount={0.bankaccount}, '
            'entity={0.entity}, order_by={0.order_by}, '
            'start_date={0.start_date}, end_date={0.end_date}'
            .format(self)
        )

        if self.entity:
            txs = models.db.session.query(models.Transaction) \
                .join(models.Account,
                      models.Transaction.account_id == models.Account.id) \
                .join(models.Entity,
                      models.Account.entity_id == models.Entity.id) \
                .filter(models.Account.entity == self.entity,
                        models.Transaction.is_deleted == 0)
        elif self.bankaccount:
            txs = models.db.session.query(models.Transaction) \
                .join(models.Account,
                      models.Transaction.account_id == models.Account.id) \
                .join(models.Entity,
                      models.Account.entity_id == models.Entity.id) \
                .filter(models.Transaction.bankaccount == self.bankaccount,
                        models.Transaction.is_deleted == 0)

        # If not transactions are found, provide empty report.
        if not txs:
            return {}

        # If order_by is invalid, provide empty report.
        if self.order_by not in ['account', 'entity', 'bankaccount', 'amount']:
            return {}

        if self.start_date:
            txs = txs.filter(models.Transaction.date >= self.start_date)

        if self.end_date:
            txs = txs.filter(models.Transaction.date <= self.end_date)

        txs = txs.all()

        income_account = models.db.session.query(models.AccountType) \
            .filter(models.AccountType.name == 'Revenue',
                    models.AccountType.parent == None).one()

        expense_account = models.db.session.query(models.AccountType) \
            .filter(models.AccountType.name == 'Expenses',
                    models.AccountType.parent == None).one()

        for tx in txs:
            app.logger.debug("Processing Transaction; id={0.id}".format(tx))
            amount = (tx.credit + tx.debit)
            account = tx.account
            bankaccount = tx.bankaccount
            accounttype = account.accounttype
            entity = account.entity

            account_info = {'account_name': account.name, 'amount': amount}
            bankaccount_info = {'bankaccount_id': bankaccount.id,
                                'bankaccount_bank': bankaccount.bank,
                                'bankaccount_number': bankaccount.number}
            entity_info = {'entity_id': entity.id, 'entity_name': entity.name}

            if accounttype.parent == income_account:
                if account.id not in self.transactions.income:
                    self.transactions.income[account.id] = (
                        merge_dicts(
                            account_info, bankaccount_info, entity_info
                        )
                    )
                else:
                    self.transactions.income[account.id]['amount'] += amount
                self.transactions.income_total_amount += amount
            elif accounttype.parent == expense_account:
                if account.id not in self.transactions.expenses:
                    self.transactions.expenses[account.id] = (
                        merge_dicts(
                            account_info, bankaccount_info, entity_info
                        )
                    )
                else:
                    self.transactions.expenses[account.id]['amount'] \
                        += amount
                self.transactions.expenses_total_amount += amount
        return {'name': self.title,
                'start_date': self.start_date,
                'end_date': self.end_date,
                'transactions': self.transactions.__dict__}


class ProfitLossTransactions():
    def __init__(self):
        self.income = {}
        self.income_total_amount = 0
        self.expenses = {}
        self.expenses_total_amount = 0
