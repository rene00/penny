import sys
import os

sys.path.insert(1, os.path.join(sys.path[0], ".."))

from app import app, models
from app.models import session
from flask_script import Manager
import pandas as pd

manager = Manager(app)


@manager.option("--account-id", dest="account_id", required=True)
@manager.option("--monthly", required=False, action="store_true")
@manager.option("--monthly-count", required=False, default=3)
@manager.option("--weekly", required=False, action="store_true")
@manager.option("--weekly-count", required=False, default=12)
@manager.option("--daily", required=False, action="store_true")
@manager.option("--daily-count", required=False, default=30)
def run(account_id, monthly, monthly_count, weekly, weekly_count, daily, daily_count):
    transactions = (
        session.query(
            models.Transaction.date, models.Transaction.debit, models.Transaction.credit
        )
        .filter(
            models.Transaction.account_id == account_id,
            models.Transaction.is_deleted == 0,
        )
        .order_by(models.Transaction.date.desc())
        .statement
    )
    pd.set_option("display.max_rows", 1000)
    df = pd.read_sql_query(
        sql=transactions,
        con=session.bind,
        parse_dates=["date"],
    )
    df["total"] = df["debit"] + df["credit"]
    df["total"] = df["total"].apply(pd.to_numeric) / float(100)
    df.pop("debit")
    df.pop("credit")
    df.date = pd.to_datetime(df.date)
    df.set_index("date", inplace=True)

    if monthly:
        print(df.resample("M").sum().tail(int(monthly_count)))
    if weekly:
        print(df.resample("W").sum().tail(int(weekly_count)))
    if daily:
        print(df.resample("D").sum().tail(int(daily_count)))


if __name__ == "__main__":
    manager.run()
