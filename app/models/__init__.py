from app.common.currency import to_cents
from datetime import datetime
from flask import url_for
from flask_security import RoleMixin, UserMixin
from flask_sqlalchemy import SQLAlchemy
from marshmallow import Schema, fields
from sqlalchemy.sql import func
import hashlib
import locale
import pytz

db = SQLAlchemy()
session = db.session

locale.setlocale(locale.LC_ALL, 'en_AU.UTF-8')


def utcnow():
    "Return a UTC datetime object with the current time."
    return datetime.utcnow().replace(tzinfo=pytz.utc)


roles_users = db.Table(
    'roles_users',
    db.Column('user_id', db.Integer(), db.ForeignKey('user.id')),
    db.Column('role_id', db.Integer(), db.ForeignKey('role.id'))
)


class Role(db.Model, RoleMixin):
    __tablename__ = 'role'

    id = db.Column(db.Integer, primary_key=True)
    name = db.Column(db.String(128), unique=True)
    description = db.Column(db.String(128), nullable=True)
    user_id = db.Column(db.Integer, db.ForeignKey('user.id'))


class User(db.Model, UserMixin):
    __tablename__ = 'user'

    id = db.Column(db.Integer, primary_key=True)
    email = db.Column(db.String(128), unique=True)
    password = db.Column(db.String(128), nullable=False)
    first_name = db.Column(db.String(1024))
    last_name = db.Column(db.String(1024))
    date_added = db.Column(db.DateTime, default=utcnow)
    active = db.Column(db.Boolean, default=False)
    confirmed_at = db.Column(db.DateTime)

    # Enabled if user is part of alpha test.
    alpha_enabled = db.Column(db.Boolean, default=True)

    entities = db.relationship('Entity', backref='user')
    bankaccounts = db.relationship('BankAccount', backref='user')
    roles = db.relationship('Role', secondary=roles_users,
                            backref=db.backref('users', lazy='dynamic'))
    transactions = db.relationship('Transaction', backref='user')
    accounts = db.relationship('Account', backref='user')
    accountmatches = db.relationship('AccountMatch', backref='user')
    transactionuploads = db.relationship('TransactionUpload', backref='user')

    def is_active(self):
        return True

    def is_authenticated(self):
        return True

    def is_anonymous(self):
        return False

    def get_id(self):
        return self.id


class EntityType(db.Model):
    __tablename__ = 'entitytype'

    id = db.Column(db.Integer, primary_key=True)
    name = db.Column(db.String(128), unique=True)

    entities = db.relationship('Entity', backref='entitytype')


class EntitySchema(Schema):
    name_as_html = fields.Method('get_name_as_html')

    class Meta:
        fields = ('id', 'name', 'name_as_html')

    def get_name_as_html(self, obj):
        return '<a href="{url}">{entity.name}</a>'.format(
            entity=obj, url=url_for('entities.entity', id=obj.id))


class Entity(db.Model):
    __tablename__ = 'entity'

    id = db.Column(db.Integer, primary_key=True)
    user_id = db.Column(db.Integer, db.ForeignKey('user.id'))
    entitytype_id = db.Column(db.Integer, db.ForeignKey('entitytype.id'))
    name = db.Column(db.String(128))

    account = db.relationship('Account', backref='entity')
    bankaccounts = db.relationship('BankAccount', backref='entity')

    def dump(self, **kwargs):
        return EntitySchema(**kwargs).dump(self).data


class BankAccountType(db.Model):
    __tablename__ = 'bankaccounttype'

    id = db.Column(db.Integer, primary_key=True)
    name = db.Column(db.String(128), unique=True)
    desc = db.Column(db.String(128))

    bankaccounts = db.relationship('BankAccount', backref='bankaccounttype')


class BankAccountSchema(Schema):

    total_balance = fields.Method('get_total_balance')
    number_as_html = fields.Method('get_number_as_html')
    bankaccounttype_desc = fields.Method('get_bankaccounttype_desc')
    total_balance_as_html = fields.Method('get_total_balance_as_html')

    class Meta:
        fields = ('id', 'user_id', 'number', 'bank', 'bankaccounttype_id',
                  'entity_id', 'desc', 'date_added', 'total_balance',
                  'number_as_html', 'bankaccounttype_desc',
                  'total_balance_as_html')

    def get_total_balance(self, obj):
        total_balance = float(obj.total_balance)
        return locale.currency(float(total_balance / float(100)),
                               grouping=True)

    def get_number_as_html(self, obj):
        return '<a href="{url}">{bankaccount.number}</a>'.format(
            bankaccount=obj,
            url=url_for('bankaccounts.bankaccount', id=obj.id))

    def get_total_balance_as_html(self, obj):
        return '<a href="{url}">{total_balance}</a>'.format(
            total_balance=self.get_total_balance(obj),
            url=url_for('transactions.bankaccount', id=obj.id))

    def get_bankaccounttype_desc(self, obj):
        name = None
        if obj.bankaccounttype:
            name = obj.bankaccounttype.desc
        return name


class BankAccount(db.Model):
    __tablename__ = 'bankaccount'

    id = db.Column(db.Integer, autoincrement=True, primary_key=True)
    user_id = db.Column(db.Integer, db.ForeignKey('user.id'), nullable=False)
    number = db.Column(db.String(128), nullable=False)
    bank = db.Column(db.String(128), nullable=True)
    bankaccounttype_id = db.Column(db.Integer,
                                   db.ForeignKey('bankaccounttype.id'),
                                   nullable=True)
    entity_id = db.Column(db.Integer, db.ForeignKey('entity.id'),
                          nullable=False)
    desc = db.Column(db.String(128), nullable=True)
    date_added = db.Column(db.DateTime, default=utcnow, nullable=False)

    accountmatches = db.relationship('AccountMatch', backref='bankaccount')
    transactions = db.relationship('Transaction', backref='bankaccount')

    def __repr__(self):
        return '<BankAccount {0.id}, {0.number}>'.format(self)

    def dump(self, **kwargs):
        return BankAccountSchema(**kwargs).dump(self).data

    @property
    def total_balance(self):
        sums = db.session.query(
            func.sum(Transaction.credit),
            func.sum(Transaction.debit)).filter(
                Transaction.bankaccount_id == self.id,
                Transaction.user_id == self.user_id,
                Transaction.is_deleted == False,
                Transaction.is_archived == False).one()
        (credit, debit) = sums
        if credit is None:
            credit = 0
        if debit is None:
            debit = 0
        return (credit + debit)


class TransactionSchema(Schema):
    date = fields.DateTime(format="%Y-%m-%d")
    bankaccount_id = fields.Method('get_bankaccount_id')
    bankaccount = fields.Nested('BankAccountSchema', allow_null=True)
    account = fields.Nested('AccountSchema', allow_null=True)
    credit = fields.Method('get_credit')
    debit = fields.Method('get_debit')
    amount = fields.Method('get_amount')
    memo_as_html = fields.Method('get_memo_as_html')
    account_name = fields.Method('get_account_name')
    account_as_html = fields.Method('get_account_as_html')

    class Meta:
        fields = (
            'id', 'date', 'memo', 'debit', 'credit', 'bankaccount',
            'fitid', 'account', 'memo_as_html', 'account_name',
            'account_as_html')

    def get_bankaccount_id(self, obj):
        "Return the bankaccount id of the transaction."
        try:
            return obj.bankaccount.id
        except:
            return None

    def get_account_name(self, obj):
        "Return the account name of the transaction."
        return obj.account.name

    def get_credit(self, obj):
        return obj._credit

    def get_debit(self, obj):
        return obj._debit

    def get_amount(self, obj):
        return obj._amount

    def get_memo_as_html(self, obj):
        return '<a href="{url}">{transaction.memo}</a>'.format(
            transaction=obj, url=url_for('transactions.transaction',
                                         id=obj.id))

    def get_account_as_html(self, obj):
        html = None
        if obj.account:
            html = ('<a href="{url}">{account.name}</a>'.
                    format(account=obj.account,
                           url=url_for('accounts.account',
                                       id=obj.account.id)))
        return (html or '')


class Transaction(db.Model):
    __tablename__ = 'transaction'

    id = db.Column(db.Integer, primary_key=True)
    user_id = db.Column(db.Integer, db.ForeignKey('user.id'))
    date = db.Column(db.DateTime, default=utcnow)
    debit = db.Column(db.Integer, nullable=True)
    credit = db.Column(db.Integer, nullable=True)
    memo = db.Column(db.String(512))
    bankaccount_id = db.Column(db.Integer, db.ForeignKey('bankaccount.id'),
                               nullable=True)
    parent_id = db.Column(db.Integer, db.ForeignKey('transaction.id'))
    children = db.relationship('Transaction')

    account_id = db.Column(db.Integer, db.ForeignKey('account.id'))
    transaction_hash = db.Column(db.String(128), unique=True)
    fitid = db.Column(db.String(128))
    paypalid = db.Column(db.String(128))
    is_deleted = db.Column(db.Boolean, default=False)
    is_archived = db.Column(db.Boolean, default=False)
    date_added = db.Column(db.DateTime, default=utcnow)

    notes = db.relationship('TransactionNote', backref='transaction')
    attachments = db.relationship('TransactionAttachment',
                                  backref='transaction')

    def __str__(self):
        return """
Id: {0.id}
Date: {0.date}
Memo: {0.memo}
Debit: {0.debit}
Credit: {0.credit}
Hash: {0.transaction_hash}
""".format(self)

    def dump(self):
        "Return serialized object."
        return TransactionSchema().dump(self).data

    @property
    def _credit(self):
        return locale.currency(float(self.credit / float(100)), grouping=True)

    @property
    def _debit(self):
        return locale.currency(float(self.debit / float(100)), grouping=True)

    @property
    def _amount(self):
        """ Return amount.

        The amount is returned as a float in locale.
        """
        return float(
            locale.currency(
                float((self.debit + self.credit) / float(100)),
                grouping=True, symbol=False
            ).replace(',', '')
        )

    def set_amount(self, new_amount):
        """Save amount as credit or debit.

        The amount value will come from the transaction form as a float.
        It will be saved as credit or debit depending on whether its
        postive or negative.

        Returns:
            obj: self.
        """
        new_amount = to_cents(new_amount)
        self.debit = self.credit = 0
        if new_amount > 0:
            self.credit = new_amount
        else:
            self.debit = new_amount
        return self


class TransactionNote(db.Model):
    __tablename__ = 'transactionnote'

    id = db.Column(db.Integer, primary_key=True)
    note = db.Column(db.String(1024), nullable=True)
    date_added = db.Column(db.DateTime, default=utcnow)
    transaction_id = db.Column(db.Integer, db.ForeignKey('transaction.id'))


class TransactionAttachment(db.Model):
    __tablename__ = 'transactionattachment'

    id = db.Column(db.Integer, primary_key=True)
    name = db.Column(db.String(1024), nullable=True)
    desc = db.Column(db.String(1024), nullable=True)
    filename = db.Column(db.String(1024), nullable=False)
    filepath = db.Column(db.String(1024), nullable=False)
    attachment_hash = db.Column(db.String(1024), nullable=False)
    transaction_id = db.Column(db.Integer, db.ForeignKey('transaction.id'))
    date_added = db.Column(db.DateTime, default=utcnow)


class TransactionUpload(db.Model):
    __tablename__ = 'transactionupload'

    id = db.Column(db.Integer, primary_key=True)
    user_id = db.Column(db.Integer, db.ForeignKey('user.id'))
    filename = db.Column(db.String(1024), nullable=True)
    filepath = db.Column(db.String(1024), nullable=False)
    upload_hash = db.Column(db.String(1024), nullable=False)
    date_added = db.Column(db.DateTime, default=utcnow)


class AccountType(db.Model):
    __tablename__ = 'accounttype'

    id = db.Column(db.Integer, primary_key=True)
    name = db.Column(db.String(128))
    parent_id = db.Column(db.Integer, db.ForeignKey('accounttype.id'),
                          index=True)
    date_added = db.Column(db.DateTime, default=utcnow)

    parent = db.relationship(lambda: AccountType, remote_side=id)
    accounts = db.relationship('Account', backref='accounttype')


class AccountSchema(Schema):
    name_as_html = fields.Method('get_name_as_html')
    entity_name = fields.Method('get_entity_name')
    account_type_name = fields.Method('get_account_type_name')

    class Meta:
        fields = ('id', 'name', 'name_as_html', 'entity_name',
                  'account_type_name')

    def get_entity_name(self, obj):
        return '{0.entity.name}'.format(obj)

    def get_account_type_name(self, obj):
        return '{0.accounttype.name}'.format(obj)

    def get_name_as_html(self, obj):
        return '<a href="{url}">{account.name}</a>'.format(
            account=obj, url=url_for('accounts.account', id=obj.id))


class Account(db.Model):
    __tablename__ = 'account'

    id = db.Column(db.Integer, primary_key=True)
    code = db.Column(db.String(128))
    name = db.Column(db.String(128))
    desc = db.Column(db.String(128))
    user_id = db.Column(db.Integer, db.ForeignKey('user.id'))
    accounttype_id = db.Column(db.Integer, db.ForeignKey('accounttype.id'))
    entity_id = db.Column(db.Integer, db.ForeignKey('entity.id'),
                          nullable=True)
    # parent = ForeignKeyField('self', null=True)
    # is_transfer = BooleanField(default=True)
    date_added = db.Column(db.DateTime, default=utcnow)

    accountmatches = db.relationship('AccountMatch', backref='account')
    transactions = db.relationship('Transaction', backref='account')

    def dump(self, **kwargs):
        return AccountSchema(**kwargs).dump(self).data


class AccountMatchSchema(Schema):
    name_as_html = fields.Method('get_name_as_html')
    account_name = fields.Method('get_account_name')
    entity_name = fields.Method('get_entity_name')

    class Meta:
        fields = ('id', 'name', 'name_as_html', 'desc', 'account_name',
                  'entity_name')

    def get_account_name(self, obj):
        return '{0.account.name}'.format(obj)

    def get_entity_name(self, obj):
        return '{0.account.entity.name}'.format(obj)

    def get_name_as_html(self, obj):
        return '<a href="{url}">{accountmatch.name}</a>'.format(
            accountmatch=obj, url=url_for('accountmatches.accountmatch',
                                          id=obj.id))


class AccountMatch(db.Model):
    __tablename__ = 'accountmatch'

    id = db.Column(db.Integer, primary_key=True)
    user_id = db.Column(db.Integer, db.ForeignKey('user.id'))
    name = db.Column(db.String(128))
    desc = db.Column(db.String(128))
    account_id = db.Column(db.Integer, db.ForeignKey('account.id'))
    bankaccount_id = db.Column(db.Integer, db.ForeignKey('bankaccount.id'),
                               nullable=True)
    date_added = db.Column(db.DateTime, default=utcnow)

    accountmatchfilterregexes = db.relationship('AccountMatchFilterRegex',
                                                backref='accountmatch')

    def dump(self, **kwargs):
        return AccountMatchSchema(**kwargs).dump(self).data


class AccountMatchFilterRegex(db.Model):
    __tablename__ = 'accountmatchfilterregex'

    id = db.Column(db.Integer, primary_key=True)
    regex = db.Column(db.String(128))
    accountmatch_id = db.Column(db.Integer, db.ForeignKey('accountmatch.id'))
    date_added = db.Column(db.DateTime, default=utcnow)
