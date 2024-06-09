from sqlalchemy.sql import func
from penny import models
import calendar
import datetime
from dateutil.rrule import rrule, MONTHLY
from typing import List, TypeVar


class Month:
    def __init__(self, year, month, amount, **kwargs):
        self.year = year
        self.month = month
        self.amount = amount
        self.previousMonth = kwargs.get("previousMonth", 0)
        self.avg = kwargs.get("avg", 0)
        self.previousAvg = kwargs.get("previousAvg", 0)

    @property
    def daycount(self):
        return calendar.monthrange(int(self.year), int(self.month))[1]

    @property
    def date(self):
        return "{0}-{1}".format(self.year, self.month)

    @property
    def change(self):
        return self.amount - self.previousMonth

    def changeAvg(self):
        return self.avg - self.previousAvg


AccountOrTag = TypeVar("AccountOrTag", models.Account, models.Tag)


class ReportsMonthlyBreakdown:
    def __init__(self, account_or_tag: AccountOrTag):
        self.account_or_tag: AccountOrTag = account_or_tag
        self._allAmounts: List[int] = []

    @property
    def _firstTransaction(self):
        q = (
            models.db.session.query(models.Transaction)
            .filter(
                models.Transaction.is_deleted == False,
                models.Transaction.is_archived == False,
                models.Transaction.user_id == self.account_or_tag.user.id,
            )
            .order_by(models.Transaction.date)
        )

        if isinstance(self.account_or_tag, models.Account):
            q = q.filter(models.Transaction.account_id == self.account_or_tag.id)
        else:
            q = q.filter(
                models.Transaction.tags.any(models.Tag.id == self.account_or_tag.id)
            )

        return q.first()

    @property
    def _totalDays(self):
        # Only go back 5 years for now on this report. Maybe this report could accept start and end date parameters.
        delta = datetime.datetime.now() - self._firstTransaction.date
        days = delta.days
        if days > 1825:
            days = 1825
        return days

    def generate(self):
        report = {"transactions": {}}
        data = {}

        if not self._firstTransaction:
            return report

        end_date = datetime.datetime.now()
        start_date = end_date - datetime.timedelta(days=self._totalDays)
        start_date = start_date.replace(day=1)

        if models.db.engine.name == "sqlite":
            q = (
                models.db.session.query(
                    func.strftime("%Y", models.Transaction.date).label("year"),
                    func.strftime("%m", models.Transaction.date).label("month"),
                    func.sum(models.Transaction.credit).label("credit"),
                    func.sum(models.Transaction.debit).label("debit"),
                )
                .filter(
                    models.Transaction.is_deleted == False,  # noqa[W0612]
                    models.Transaction.is_archived == False,
                    models.Transaction.user_id == self.account_or_tag.user.id,
                    models.Transaction.date >= start_date,
                    models.Transaction.date <= end_date,
                )
                .group_by(func.strftime("%Y-%m-01", models.Transaction.date))
                .order_by(models.Transaction.date)
            )
        else:
            q = (
                models.db.session.query(
                    func.date_format(models.Transaction.date, "%Y").label("year"),
                    func.date_format(models.Transaction.date, "%m").label("month"),
                    func.sum(models.Transaction.credit).label("credit"),
                    func.sum(models.Transaction.debit).label("debit"),
                )
                .filter(
                    models.Transaction.is_deleted == False,  # noqa[W0612]
                    models.Transaction.is_archived == False,
                    models.Transaction.user_id == self.account_or_tag.user.id,
                    models.Transaction.date >= start_date,
                    models.Transaction.date <= end_date,
                )
                .group_by(func.date_format(models.Transaction.date, "%Y-%m-01"))
                .order_by(models.Transaction.date)
            )

        if isinstance(self.account_or_tag, models.Account):
            q = q.filter(models.Transaction.account_id == self.account_or_tag.id)
        else:
            q = q.filter(
                models.Transaction.tags.any(models.Tag.id == self.account_or_tag.id)
            )

        for transaction in q.all():
            month = Month(
                year=transaction.year,
                month=transaction.month,
                amount=int(transaction.credit + transaction.debit),
            )
            data[month.date] = month

        count = 0
        for d in rrule(freq=MONTHLY, dtstart=start_date, until=end_date):
            month = Month(
                year=d.strftime("%Y"),
                month=d.strftime("%m"),
                amount=0,
                previousMonth=0,
                avg=0,
            )

            exists = data.get(month.date)
            if exists:
                month.amount = exists.amount

            self._allAmounts.append(month.amount)
            if count >= 1:
                month.previousMonth = self._allAmounts[count - 1]

            lastAmounts = self._allAmounts[-12:]
            try:
                month.avg = sum(lastAmounts) / len(lastAmounts)
            except ZeroDivisionError:
                pass

            report["transactions"][month.date] = month
            count += 1

        return report
