from penny import models, util
from penny.common import forms
from flask import Blueprint, g, render_template, url_for, redirect
from flask_security import login_required
from sqlalchemy.orm.exc import NoResultFound
from sqlalchemy.sql import func
from penny.resources.accounts.forms import FormAccount


accounts = Blueprint("accounts", __name__)


@accounts.route("/accounts")
@login_required
def _accounts():
    return render_template("accounts.html", data_url=url_for("data_accounts.accounts"))


@accounts.route(
    "/accounts/<int:id>",
    defaults={"start_date": None, "end_date": None},
    methods=["GET", "POST"],
)
@accounts.route(
    "/accounts/<int:id>/<string:start_date>/<string:end_date>", methods=["GET", "POST"]
)
@login_required
def account(id, start_date, end_date):

    try:
        account = (
            models.db.session.query(models.Account).filter_by(id=id, user=g.user).one()
        )
    except NoResultFound:
        return url_for("accounts._accounts")

    form = FormAccount(obj=account)
    form.accounttype.choices = forms.get_accounttype_as_choices()
    form.entity.choices = forms.get_entities_as_choices()

    if form.validate_on_submit():
        account.name = form.name.data
        account.desc = form.desc.data

        if form.accounttype.data:
            account.accounttype_id = form.accounttype.data

        if form.entity.data:
            account.entity_id = form.entity.data

        models.db.session.add(account)
        models.db.session.commit()

    form.set_defaults(account)

    transactions = models.db.session.query(
        func.sum(models.Transaction.credit).label("credit"),
        func.sum(models.Transaction.debit).label("debit"),
    ).filter(
        models.Transaction.is_deleted == False,  # noqa[W0612]
        models.Transaction.is_archived == False,
        models.Transaction.account_id == account.id,
        models.Transaction.user_id == g.user.id,
    )

    transactions_amount = 0
    for transaction in transactions.all():
        amount = 0
        if transaction.credit:
            amount = transaction.credit
        if transaction.debit:
            amount += transaction.debit
        transactions_amount += amount

    return render_template(
        "account.html",
        form=form,
        account=account,
        transactions_amount=util.convert_to_float(int(transactions_amount)),
    )


@accounts.route("/accounts/add", methods=["GET", "POST"])
@login_required
def add():
    form = FormAccount()
    form.accounttype.choices = forms.get_accounttype_as_choices()
    form.entity.choices = forms.get_entities_as_choices()
    account = None

    if form.validate_on_submit():
        account = models.Account(user_id=g.user.id)
        account.name = form.name.data
        account.desc = form.desc.data

        if form.accounttype.data:
            account.accounttype_id = form.accounttype.data

        if form.entity.data:
            account.entity_id = form.entity.data

        models.db.session.add(account)
        models.db.session.commit()

        return redirect(url_for("accounts.account", id=account.id))

    return render_template("account.html", form=form, account=account)
