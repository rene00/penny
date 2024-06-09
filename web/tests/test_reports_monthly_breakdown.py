from penny.resources.reports import ReportsMonthlyBreakdown
from .conftest import (
    create_entity,
    create_bankaccount,
    create_account,
    create_transaction,
)
from penny import models


def test_reports_monthly_breakdown(client, auth):
    with client:
        auth.login()
        create_entity(client)
        create_bankaccount(client)
        create_account(client)
        create_transaction(client, memo="test1", credit=1, account=1)
        create_transaction(client, memo="test2", credit=2, account=1)
        account = models.db.session.query(models.Account).one()

        report = ReportsMonthlyBreakdown(account).generate()
        transactions = report.get("transactions", False)
        assert transactions

        month = next(iter(transactions.values()))
        assert month.year
        assert month.month
        assert month.amount
