from flask import current_app as app, g     # noqa[W0611]
from sqlalchemy.sql import func
from penny import models, util                # noqa[E402]
import calendar
import datetime
from dateutil.rrule import rrule, MONTHLY


class Month:
    def __init__(self, year, month, amount):
        self.year = year
        self.month = month
        self.amount = amount

    @property
    def daycount(self):
        return calendar.monthrange(int(self.year), int(self.month))[1]

    @property
    def date(self):
        return '{0}-{1}'.format(self.year, self.month)


class ReportsMonthlyBreakdown:
    def __init__(self, account):
        self.account = account

    def generate(self):
        report = {'transactions': {}}
        data = {}

        end_date = datetime.datetime.now()
        start_date = end_date - datetime.timedelta(days=1095)
        start_date = start_date.replace(day=1)

        transactions = models.db.session.query(
                func.date_format(models.Transaction.date, '%Y').label("year"),
                func.date_format(models.Transaction.date, '%m').label("month"),
                func.sum(models.Transaction.credit).label("credit"),
                func.sum(models.Transaction.debit).label("debit"),
            ) \
            .filter(
                    models.Transaction.is_deleted == False,  # noqa[W0612]
                    models.Transaction.is_archived == False,
                    models.Transaction.account_id == self.account.id,
                    models.Transaction.user_id == g.user.id,
                    models.Transaction.date >= start_date,
                    models.Transaction.date <= end_date,
                ) \
            .group_by(func.date_format(models.Transaction.date, '%Y-%m-01')) \
            .order_by(models.Transaction.date)

        for transaction in transactions.all():
            amount = util.convert_to_float(
                int(transaction.credit + transaction.debit)
            )
            month = Month(
                year=transaction.year, month=transaction.month, amount=amount
            )
            data[month.date] = month

        for d in rrule(freq=MONTHLY, dtstart=start_date, until=end_date):
            month = Month(
                year=d.strftime("%Y"), month=d.strftime("%m"), amount="$0.00"
            )
            exists = data.get(month.date)
            if exists:
                month.amount = exists.amount
            report['transactions'][month.date] = month

        return report
