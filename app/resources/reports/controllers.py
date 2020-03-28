from app import models, util
from app.resources.reports import ReportsProfitLoss, ReportsSavingsRate
from app.common import forms
from datetime import datetime as dt
from flask import Blueprint, g, render_template, url_for, make_response
from flask_security import login_required
from sqlalchemy.orm.exc import NoResultFound
from sqlalchemy.sql import func
from flask_wtf import Form
from wtforms import DateField, SelectField
from dateutil.rrule import rrule, MONTHLY
import datetime
import calendar

reports = Blueprint('reports', __name__, url_prefix='/reports')


@reports.route('/')
@login_required
def _reports():
    return render_template(
        'reports.html', data_url=url_for('data_reports.reports')
    )


DATE_FMT = '%Y%m%d'
DELTA_DAYS = 365


class FormMonthlyBreakdown(Form):
    now = datetime.datetime.now()
    delta = datetime.timedelta(days=DELTA_DAYS)

    account = SelectField(u'Account', validators=[], coerce=int)

    datepicker_start = DateField(
        'Start Date', format='%d/%m/%Y', validators=[],
        default=(now - delta))

    datepicker_end = DateField(
        'End Date', format='%d/%m/%Y', validators=[],
        default=now)

    def get_account(self):
        """Return the account."""
        try:
            account = models.db.session.query(models.Account) \
                .filter_by(id=self.account.data, user=g.user).one()
        except NoResultFound:
            account = None
        finally:
            return account

    def get_start(self, fmt=DATE_FMT):
        return self.datepicker_start.data.strftime(fmt)

    def get_end(self, fmt=DATE_FMT):
        return self.datepicker_end.data.strftime(fmt)


class FormBasicDates(Form):

    now = datetime.datetime.now()
    delta = datetime.timedelta(days=DELTA_DAYS)

    datepicker_start = DateField(
        'Start Date', format='%d/%m/%Y', validators=[],
        default=(now - delta))

    datepicker_end = DateField(
        'End Date', format='%d/%m/%Y', validators=[],
        default=now)

    def get_start(self, fmt=DATE_FMT):
        return self.datepicker_start.data.strftime(fmt)

    def get_end(self, fmt=DATE_FMT):
        return self.datepicker_end.data.strftime(fmt)


class MonthlyBreakdownMonth:
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


def generate_report_account_monthly(account, start_date, end_date):
    report = {'transactions': {}}
    data = {}

    transactions = models.db.session.query(
            func.date_format(models.Transaction.date, '%Y').label("year"),
            func.date_format(models.Transaction.date, '%m').label("month"),
            func.sum(models.Transaction.credit).label("credit"),
            func.sum(models.Transaction.debit).label("debit"),
        ) \
        .filter(
                models.Transaction.is_deleted == False,  # noqa[W0612]
                models.Transaction.is_archived == False,
                models.Transaction.account_id == account.id,
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
        month = MonthlyBreakdownMonth(
            year=transaction.year, month=transaction.month, amount=amount
        )
        data[month.date] = month

    for d in rrule(freq=MONTHLY, dtstart=start_date, until=end_date):
        month = MonthlyBreakdownMonth(
            year=d.strftime("%Y"), month=d.strftime("%m"), amount="$0.00"
        )
        exists = data.get(month.date)
        if exists:
            month.amount = exists.amount
        report['transactions'][month.date] = month

    return report


@reports.route('/savings-rate', methods=['GET'])
@login_required
def savings_rate():
    report = ReportsSavingsRate()
    return render_template(
        'reports/savings-rate.html',
        report=report.generate(),
    )


@reports.route('/account-monthly-breakdown/', methods=['GET', 'POST'])
@login_required
def account_monthly_breakdown():

    form = FormMonthlyBreakdown()
    form.account.choices = forms.get_account_as_choices()

    if form.datepicker_start.data:
        start_date = form.datepicker_start.data
    else:
        try:
            start_date = dt.strptime(start_date, DATE_FMT)
        except ValueError:
            return make_response(
                render_template('errors/invalid_date_range.html'), 400
            )
        else:
            form.datepicker_start.data = start_date

    if form.datepicker_end.data:
        end_date = form.datepicker_end.data
    else:
        try:
            end_date = dt.strptime(end_date, DATE_FMT)
        except ValueError:
            return make_response(
                render_template('errors/invalid_date_range.html'), 400
            )
        else:
            form.datepicker_end.data = end_date

    report = {'transactions': {}}
    account_id = None
    account_name = None

    if form.validate_on_submit():
        account = form.get_account()
        account_id = account.id
        account_name = account.name
        report = generate_report_account_monthly(
            account, start_date, end_date
        )

    return render_template(
        'reports/monthly-breakdown.html', report=report, form=form,
        start_date=form.get_start(), end_date=form.get_end(),
        account_id=account_id, account_name=account_name
    )


@reports.route('/profitloss/<string:report_resource>/<int:id>',
               defaults={'start_date': None, 'end_date': None},
               methods=['GET', 'POST'])
@reports.route('/profitloss/<string:report_resource>/<int:id>/'
               '<string:start_date>/<string:end_date>',
               methods=['GET', 'POST'])
@login_required
def profitloss(report_resource, id, start_date, end_date):

    entity = bankaccount = None

    if report_resource not in ['entity', 'bankaccount']:
        return render_template('errors/invalid_report_type')

    try:
        if report_resource == 'entity':
            entity = models.db.session.query(models.Entity)\
                .filter_by(id=id, user=g.user).one()
        elif report_resource == 'bankaccount':
            bankaccount = models.db.session.query(models.BankAccount)\
                .filter_by(id=id, user=g.user).one()
    except NoResultFound:
        return url_for('reports._reports')

    form = FormBasicDates()

    # If user has included dates in URL, take the date strings and
    # convert them to datetime objs.
    if start_date and end_date:
        form.datepicker_start.data = dt.strptime(start_date, '%Y%m%d')
        form.datepicker_end.data = dt.strptime(end_date, '%Y%m%d')

    report = ReportsProfitLoss(entity=entity, bankaccount=bankaccount,
                               start_date=form.datepicker_start.data,
                               end_date=form.datepicker_end.data)

    generated_report = report.generate()

    if not generated_report:
        return render_template('reports/general/no_transactions_found.html')
    else:
        return render_template('reports/profitloss.html',
                               report=generated_report, entity=entity,
                               bankaccount=bankaccount,
                               report_resource=report_resource,
                               form=form, start_date=form.get_start(),
                               end_date=form.get_end())
