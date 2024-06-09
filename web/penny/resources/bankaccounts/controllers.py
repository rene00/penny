from penny import models
from penny.common import forms
from flask import Blueprint, g, render_template, url_for, redirect
from flask_security import auth_required
from sqlalchemy.orm.exc import NoResultFound
from penny.resources.bankaccounts.forms import FormBankAccount


bankaccounts = Blueprint("bankaccounts", __name__)


@bankaccounts.route("/bankaccounts", endpoint="_bankaccounts")
@auth_required()
def _bankaccounts():
    return render_template(
        "bankaccounts.html", data_url=url_for("data_bankaccounts.bankaccounts")
    )


@bankaccounts.route(
    "/bankaccounts/<int:id>", methods=["GET", "POST"], endpoint="bankaccount"
)
@auth_required()
def bankaccount(id):

    try:
        bankaccount = (
            models.db.session.query(models.BankAccount)
            .filter_by(id=id, user=g.user)
            .one()
        )
    except NoResultFound:
        return url_for("bankaccounts._bankaccounts")

    form = FormBankAccount(obj=bankaccount)
    form.bankaccounttype.choices = forms.get_bankaccounttype_as_choices()
    form.entity.choices = forms.get_entities_as_choices()

    if form.validate_on_submit():
        bankaccount.bank = form.bank.data
        bankaccount.number = form.number.data
        bankaccount.desc = form.desc.data

        if form.bankaccounttype.data:
            bankaccount.bankaccounttype_id = form.bankaccounttype.data

        if form.entity.data:
            bankaccount.entity_id = form.entity.data

        models.db.session.add(bankaccount)
        models.db.session.commit()

    form.set_defaults(bankaccount)

    return render_template("bankaccount.html", form=form, bankaccount=bankaccount)


@bankaccounts.route("/bankaccounts/add", methods=["GET", "POST"], endpoint="add")
@auth_required()
def add():
    form = FormBankAccount()
    form.bankaccounttype.choices = forms.get_bankaccounttype_as_choices()
    form.entity.choices = forms.get_entities_as_choices()

    if form.validate_on_submit():
        bankaccount = models.BankAccount(user_id=g.user.id)

        if form.bank.data:
            bankaccount.bank = form.bank.data

        if form.number.data:
            bankaccount.number = form.number.data

        if form.desc.data:
            bankaccount.desc = form.desc.data

        if form.bankaccounttype.data:
            bankaccount.bankaccounttype_id = form.bankaccounttype.data

        if form.entity.data:
            bankaccount.entity_id = form.entity.data

        models.db.session.add(bankaccount)
        models.db.session.commit()

        return redirect(url_for("bankaccounts.bankaccount", id=bankaccount.id))

    return render_template("bankaccount.html", form=form)
