from app import models
from app.resources.reports import (
    ReportsProfitLoss,
    ReportsSavingsRate,
    ReportsMonthlyBreakdown
)
from app.common import forms
from datetime import datetime as dt
from flask import Blueprint, g, render_template, url_for
from flask_security import login_required
from sqlalchemy.orm.exc import NoResultFound
from flask_wtf import Form
from wtforms import DateField, SelectField
import datetime

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
    account = SelectField(u'Account', validators=[], coerce=int)

    def get_account(self):
        """Return the account."""
        try:
            account = models.db.session.query(models.Account) \
                .filter_by(id=self.account.data, user=g.user).one()
        except NoResultFound:
            account = None
        finally:
            return account


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


@reports.route('/savings-rate', methods=['GET'])
@login_required
def savings_rate():
    report = ReportsSavingsRate()
    return render_template(
        'reports/savings-rate.html',
        report=report.generate(),
    )


@reports.route(
    '/account-monthly-breakdown/<int:account_id>',
    methods=['GET', 'POST']
)
@reports.route(
    '/account-monthly-breakdown/',
    defaults={'account_id': None},
    methods=['GET', 'POST']
)
@login_required
def account_monthly_breakdown(account_id):

    form = FormMonthlyBreakdown()
    form.account.choices = forms.get_account_as_choices()

    report = {'transactions': {}}
    account = None

    if account_id:
        try:
            account = models.db.session.query(models.Account) \
                .filter_by(id=account_id, user=g.user).one()
        except NoResultFound:
            return url_for('reports.account_monthly_breakdown')
        report = ReportsMonthlyBreakdown(account).generate()

    if form.validate_on_submit():
        account = form.get_account()
        report = ReportsMonthlyBreakdown(account).generate()

    return render_template(
        'reports/monthly-breakdown.html', report=report, form=form,
        account=account
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
