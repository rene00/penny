from flask.cli import AppGroup
from penny import models
from penny.resources.reports import ReportsMonthlyBreakdown
import click

report_cli: AppGroup = AppGroup("report")


@report_cli.command("account_monthly_breakdown")
@click.argument('user_id')
@click.argument('account_id')
def report_account_monthly_breakdown(user_id: int, account_id: int) -> None:
    user: models.User = models.User.query.filter_by(id=user_id).one()
    account: models.Account = models.Account.query.filter_by(id=account_id, user=user).one()
    report = ReportsMonthlyBreakdown(account).generate()
    for k, v in report["transactions"].items():
        month = k
        amount = float(abs(v.amount) / float(100))
        print(month, amount)
    return None


@report_cli.command("tag_monthly_breakdown")
@click.argument('user_id')
@click.argument('tag_id')
def report_tag_monthly_breakdown(user_id: int, tag_id: int) -> None:
    user: models.User = models.User.query.filter_by(id=user_id).one()
    tag: models.Tag = models.Tag.query.filter_by(id=tag_id, user=user).one()
    report = ReportsMonthlyBreakdown(tag).generate()
    for k, v in report["transactions"].items():
        month = k
        amount = float(abs(v.amount) / float(100))
        avg = float(abs(v.avg) / float(100))
        print(f"{month}, {amount}, {avg:.2f}")
    return None
