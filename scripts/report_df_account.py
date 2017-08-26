import sys
import os
sys.path.insert(1, os.path.join(sys.path[0], '..'))

from app import app, models
from app.models import session
from flask_script import Manager
import pandas as pd

manager = Manager(app)


@manager.option('--account-id', dest='account_id', required=True)
def run(account_id):
    transactions = (
        session.query(
            models.Transaction.date,
            models.Transaction.debit,
            models.Transaction.credit
        ).
        filter(
            models.Transaction.account_id == account_id,
            models.Transaction.is_deleted == 0).
        order_by(
            models.Transaction.date.desc()
        ).statement
    )
    pd.set_option('display.max_rows', 1000)
    df = pd.read_sql_query(
        sql=transactions,
        con=session.bind,
        parse_dates=['date'],
    )
    df['total'] = df['debit'] + df['credit']
    df['total'] = df['total'].apply(pd.to_numeric) / float(100)
    df.pop('debit')
    df.pop('credit')
    df.date = pd.to_datetime(df.date)
    df.set_index('date', inplace=True)
    print(df.resample('M').sum().tail(3))
    print(df.resample('W').sum().tail(12))
    print(df.resample('D').sum().tail(30))

if __name__ == '__main__':
    manager.run()
