from penny import models
from penny.resources.reports import (
    ReportsProfitLoss,
    ReportsSavingsRate,
    ReportsMonthlyBreakdown,
)
from penny.common import forms
from datetime import datetime as dt
from flask import Blueprint, g, render_template, url_for, json
from flask_security.decorators import auth_required
from sqlalchemy.orm.exc import NoResultFound
from flask_wtf import FlaskForm
from wtforms import DateField, SelectField
from typing import Optional
import datetime

reports = Blueprint("reports", __name__, url_prefix="/reports")


@reports.route("/")
@auth_required()
def _reports():
    return render_template("reports.html", data_url=url_for("data_reports.reports"))


DATE_FMT = "%Y%m%d"
DELTA_DAYS = 365


class FormMonthlyBreakdown(FlaskForm):
    account = SelectField("Account", validators=[], coerce=int)

    def get_account(self):
        """Return the account."""
        account = None
        try:
            account = (
                models.db.session.query(models.Account)
                .filter_by(id=self.account.data, user=g.user)
                .one()
            )
        except NoResultFound:
            pass
        return account


class FormTagMonthlyBreakdown(FlaskForm):
    tag = SelectField("Tag", validators=[], coerce=int)

    def get_tag(self):
        return models.Tag.query.filter_by(id=self.tag.data, user=g.user).one_or_none()


class FormBasicDates(FlaskForm):

    now = datetime.datetime.now()
    delta = datetime.timedelta(days=DELTA_DAYS)

    datepicker_start = DateField(
        "Start Date", format="%d/%m/%Y", validators=[], default=(now - delta)
    )

    datepicker_end = DateField(
        "End Date", format="%d/%m/%Y", validators=[], default=now
    )

    def get_start(self, fmt=DATE_FMT):
        return self.datepicker_start.data.strftime(fmt)

    def get_end(self, fmt=DATE_FMT):
        return self.datepicker_end.data.strftime(fmt)


@reports.route("/savings-rate", methods=["GET"])
@auth_required()
def savings_rate():
    report = ReportsSavingsRate()
    return render_template(
        "reports/savings-rate.html",
        report=report.generate(),
    )


@reports.route("/account-monthly-breakdown/<int:account_id>", methods=["GET", "POST"])
@reports.route(
    "/account-monthly-breakdown/",
    defaults={"account_id": None},
    methods=["GET", "POST"],
)
@auth_required()
def account_monthly_breakdown(account_id: int) -> str:

    form = FormMonthlyBreakdown()
    form.account.choices = forms.get_account_as_choices()

    report: dict = {"transactions": {}}
    account: Optional[models.Account] = None

    if account_id:
        account = (
            models.db.session.query(models.Account)
            .filter_by(id=account_id, user=g.user)
            .one_or_none()
        )
        if account:
            report = ReportsMonthlyBreakdown(account).generate()
        else:
            return url_for("reports.account_monthly_breakdown")

    if form.validate_on_submit():
        account = form.get_account()
        if account:
            report = ReportsMonthlyBreakdown(account).generate()

    labels = []
    data = []
    for k, v in report["transactions"].items():
        labels.append(k)
        data.append(float(abs(v.amount) / float(100)))

    kwargs = {
        "report": report,
        "form": form,
        "labels": json.dumps(labels),
        "data": json.dumps(data),
    }
    if account:
        kwargs["account"] = account

    return render_template("reports/account-monthly-breakdown.html", **kwargs)


@reports.route("/tag-monthly-breakdown/<int:tag_id>", methods=["GET", "POST"])
@reports.route(
    "/tag-monthly-breakdown/",
    defaults={"tag_id": None},
    methods=["GET", "POST"],
)
@auth_required()
def tag_monthly_breakdown(tag_id: int) -> str:

    form = FormTagMonthlyBreakdown()
    form.tag.choices = forms.get_tag_as_choices()

    report: dict = {"transactions": {}}
    tag: Optional[models.Tag] = None

    if tag_id:
        tag = models.Tag.query.filter_by(id=tag_id, user=g.user).one_or_none()
        if tag:
            report = ReportsMonthlyBreakdown(tag).generate()
        else:
            return url_for("reports.tag_monthly_breakdown")

    if form.validate_on_submit():
        tag = form.get_tag()
        if tag:
            report = ReportsMonthlyBreakdown(tag).generate()

    labels: list = []
    data: list = []
    for k, v in report["transactions"].items():
        labels.append(k)
        data.append(float(abs(v.amount) / float(100)))

    kwargs = {
        "report": report,
        "form": form,
        "labels": json.dumps(labels),
        "data": json.dumps(data),
    }
    if tag:
        kwargs["tag"] = tag

    return render_template("reports/tag-monthly-breakdown.html", **kwargs)


@reports.route(
    "/profitloss/<string:report_resource>/<int:id>",
    defaults={"start_date": None, "end_date": None},
    methods=["GET", "POST"],
)
@reports.route(
    "/profitloss/<string:report_resource>/<int:id>/"
    "<string:start_date>/<string:end_date>",
    methods=["GET", "POST"],
)
@auth_required()
def profitloss(report_resource, id, start_date, end_date):

    entity = bankaccount = None

    if report_resource not in ["entity", "bankaccount"]:
        return render_template("errors/invalid_report_type")

    try:
        if report_resource == "entity":
            entity = (
                models.db.session.query(models.Entity)
                .filter_by(id=id, user=g.user)
                .one()
            )
        elif report_resource == "bankaccount":
            bankaccount = (
                models.db.session.query(models.BankAccount)
                .filter_by(id=id, user=g.user)
                .one()
            )
    except NoResultFound:
        return url_for("reports._reports")

    form = FormBasicDates()

    # If user has included dates in URL, take the date strings and
    # convert them to datetime objs.
    if start_date and end_date:
        form.datepicker_start.data = dt.strptime(start_date, "%Y%m%d")
        form.datepicker_end.data = dt.strptime(end_date, "%Y%m%d")

    report = ReportsProfitLoss(
        entity=entity,
        bankaccount=bankaccount,
        start_date=form.datepicker_start.data,
        end_date=form.datepicker_end.data,
    )

    generated_report = report.generate()

    if not generated_report:
        return render_template("reports/general/no_transactions_found.html")
    else:
        return render_template(
            "reports/profitloss.html",
            report=generated_report,
            entity=entity,
            bankaccount=bankaccount,
            report_resource=report_resource,
            form=form,
            start_date=form.get_start(),
            end_date=form.get_end(),
        )
