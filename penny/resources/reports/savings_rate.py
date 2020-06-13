from flask import current_app as app, g     # noqa[W0611]
from sqlalchemy.sql import func
from penny import models                      # noqa[E402]
import calendar
import datetime
import dateutil.relativedelta


class Month:
    def __init__(
        self,
        month,
        year,
        revenue=0,
        expenses=0,
        liabilities=0,
        **kwargs
    ):
        self.month = month
        self.year = year
        self.revenue = revenue
        self.expenses = expenses
        self.liabilities = liabilities
        self.previous = kwargs.get('previous', None)

    @property
    def daycount(self):
        return calendar.monthrange(int(self.year), int(self.month))[1]

    @property
    def outgoing(self):
        return abs(self.expenses) + abs(self.liabilities)

    @property
    def savings(self):
        return self.revenue - self.outgoing

    @property
    def ratio(self):
        _ratio = 0
        try:
            _ratio = round(self.savings / self.revenue * 100, 2)
        except ZeroDivisionError:
            pass
        return _ratio

    @property
    def ratio_change(self):
        return round((self.ratio - self.previous.ratio), 2)

    @property
    def savings_change(self):
        if self.previous:
            previous_savings = self.previous.savings
        else:
            previous_savings = 0
        return self.savings - previous_savings


class ReportsSavingsRate:
    def __init__(self):
        self.end_date = datetime.datetime.now()
        start_date = self.end_date - datetime.timedelta(days=1095)
        self.start_date = start_date.replace(day=1)

    def generate(self):
        revenue_account = models.db.session.query(models.AccountType) \
            .filter(models.AccountType.name == 'Revenue',
                    models.AccountType.parent == None).one()    # noqa[E711]

        expenses_account = models.db.session.query(models.AccountType) \
            .filter(models.AccountType.name == 'Expenses',
                    models.AccountType.parent == None).one()    # noqa[E711]

        liabilities_account = models.db.session.query(models.AccountType) \
            .filter(models.AccountType.name == 'Liabilities',
                    models.AccountType.parent == None).one()    # noqa[E711]

        transactions = models.db.session.query(
                models.Transaction,
                func.date_format(models.Transaction.date, '%Y').label("year"),
                func.date_format(models.Transaction.date, '%m').label("month"),
            ) \
            .join(models.Account,
                models.Transaction.account_id == models.Account.id) \
            .filter(
                    models.Transaction.is_deleted == False,     # noqa[W0612]
                    models.Transaction.is_archived == False,
                    models.Transaction.date >= self.start_date,
                    models.Transaction.date <= self.end_date,
                    models.Transaction.account_id != False,
                    models.Transaction.user == g.user
                ) \
            .order_by(models.Transaction.date)

        data = {}

        for d in transactions.all():
            transaction = d[0]
            transaction_year = d[1]
            transaction_month = d[2]
            year_month = "{0}-{1}".format(transaction_year, transaction_month)
            amount = (transaction.credit + transaction.debit)
            account = transaction.account

            month = data.get(
                year_month,
                Month(transaction_month, transaction_year)
            )

            # Transaction must be revenue, expense or liability.
            if account.accounttype.parent == revenue_account:
                month.revenue += amount
            elif account.accounttype.parent == expenses_account:
                month.expenses += amount
            elif account.accounttype.parent == liabilities_account:
                month.liabilities += amount
            else:
                raise Exception(
                    "transaction not revenue, expense or liability"
                )

            data[year_month] = month

        data2 = data.copy()

        # Add previous month data
        for month, month_data in data2.items():
            month_dt = datetime.datetime.strptime(month, '%Y-%m')
            previous_month = (
                month_dt + dateutil.relativedelta.relativedelta(months=-1)
            )
            previous_month_data = data.get(
                previous_month.strftime('%Y-%m'), None
            )
            month_data.previous = previous_month_data
            data[month] = month_data

        return data
