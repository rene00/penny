import sys
import os
sys.path.insert(1, os.path.join(sys.path[0], '..'))

from flask.ext.script import Manager
from app import app, models
from app.resources.reports import ReportsProfitLoss
from pprint import pprint

manager = Manager(app)
app.config['FLASK_LOG_LEVEL'] = 'DEBUG'


@manager.option('--entity-id', dest='entity_id')
@manager.option('--bankaccount-id', dest='bankaccount_id')
@manager.option('--start-date', dest='start_date')
@manager.option('--end-date', dest='end_date')
def report(**kwargs):

    entity = bankaccount = start_date = end_date = None

    entity_id = kwargs.get('entity_id')
    bankaccount_id = kwargs.get('bankaccount_id')
    start_date = kwargs.get('start_date')
    end_date = kwargs.get('end_date')

    if entity_id:
        entity = models.Entity.query.get(entity_id)
    elif bankaccount_id:
        bankaccount = models.BankAccount.query.get(bankaccount_id)

    report = ReportsProfitLoss(entity=entity, bankaccount=bankaccount,
                               start_date=start_date, end_date=end_date)
    pprint(report.generate(), indent=4)

if __name__ == '__main__':
    manager.run()
