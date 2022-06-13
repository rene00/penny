from penny import models
from flask import g
from flask_wtf import FlaskForm
from werkzeug.datastructures import MultiDict
from wtforms import (
    TextAreaField,
    FileField,
    SelectField,
    DateField,
    DecimalField,
    validators,
)
from datetime import datetime
from penny.common.currency import to_cents
from sqlalchemy.orm.exc import NoResultFound
from wtforms_sqlalchemy.fields import QuerySelectField


def get_account_label(obj):
    return "{0.entity.name} - {0.name}".format(obj)


def get_bankaccount_label(obj):
    return "{0.bank} - {0.number}".format(obj)

def get_tag_label(obj):
    return "{0.name} - {0.desc}".format(obj)


class FormTransaction(FlaskForm):
    account = QuerySelectField(
        "account",
        get_label=get_account_label,
        allow_blank=True,
        validators=[],
        get_pk=lambda a: a.id,
    )
    bankaccount = QuerySelectField(
        "bankaccount",
        get_label=get_bankaccount_label,
        allow_blank=True,
        validators=[],
        get_pk=lambda b: b.id,
    )
    tags = QuerySelectField(
        "tags",
        get_label=get_tag_label,
        allow_blank=True,
        validators=[],
        get_pk=lambda b: b.id,
    )
    attachment = FileField(u"Filename", validators=[])
    note = TextAreaField(u"Note", default="", validators=[])
    amount = DecimalField(u"Amount", validators=[])
    id = None

    def reset(self):
        blankdata = MultiDict([])
        self.process(blankdata)

    def set_defaults(self, transaction):
        """Set default values for resources of the form based off
        transaction."""


        if transaction.account:
            self.account.default = transaction.account.id

        if transaction.bankaccount:
            self.bankaccount.default = transaction.bankaccount.id

    def set_data(self, transaction):
        """Set the data attribute for each field based on transaction."""
        self.id = transaction.id
        if transaction._amount:
            self.amount.data = transaction._amount


class FormTransactionAdd(FlaskForm):
    date = DateField(
        u"Date", default=datetime.now(), validators=[validators.DataRequired()]
    )
    debit = DecimalField(u"Debit", default=0, validators=[])
    credit = DecimalField(u"Credit", default=0, validators=[])
    memo = TextAreaField(u"Memo", default="", validators=[validators.DataRequired()])
    account = SelectField(u"Account", validators=[], coerce=int)
    bankaccount = SelectField(u"Bank Account", validators=[], coerce=int)

    def get_credit(self):
        """Return credit."""
        return to_cents(self.credit.data)

    def get_debit(self):
        """Return debit.

        Notes:
            The debit amount should always be negative.
        """
        return -abs(to_cents(self.debit.data))

    def get_account(self):
        """Return the account."""
        try:
            account = (
                models.db.session.query(models.Account)
                .filter_by(id=self.account.data, user=g.user)
                .one()
            )
        except NoResultFound:
            account = None
        finally:
            return account

    def get_bankaccount(self):
        """Return the bankaccount."""
        try:
            bankaccount = (
                models.db.session.query(models.BankAccount)
                .filter_by(id=self.bankaccount.data, user=g.user)
                .one()
            )
        except NoResultFound:
            bankaccount = None
        finally:
            return bankaccount

    def get_tags(self):
        return models.db.session.query(models.Tag).filter_by(transaction_id=self.transaction.data).all()


class FormTransactionSplit(FlaskForm):
    split_amount = TextAreaField(
        u"Amount", default="", validators=[validators.DataRequired()]
    )
    split_memo = TextAreaField(
        u"Memo", default="", validators=[validators.DataRequired()]
    )

    split_account = QuerySelectField(
        "account",
        get_label=get_account_label,
        allow_blank=True,
        validators=[],
        get_pk=lambda a: a.id,
    )

    def reset(self):
        # XXX: use reset_csrf() here. See
        # https://gist.github.com/tomekwojcik/953046
        blankdata = MultiDict([])
        self.process(blankdata)

    def get_amount(self):
        amount = to_cents(self.amount.data)
        if amount < 0:
            credit = 0
            debit = amount
        else:
            credit = amount
            debit = 0

        return (credit, debit)

    def get_account(self):
        if hasattr(self.account, "data") and self.account.data != 0:
            return self.account.data
        else:
            return None


class FormTransactionUpload(FlaskForm):
    upload = FileField(u"Filename", validators=[validators.DataRequired()])
