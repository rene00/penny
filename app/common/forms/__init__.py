from app import models
from flask import g


def get_accounttype_as_choices(first_choice=None):
    """Get accounttypes, sort and return as form choices field.

    Note:
        Account Types with parents should always be excluded from the
        list of choices returned.
    """

    choices = [(u'0', (first_choice or ''))]
    for accounttype in (models.db.session.
                        query(models.AccountType).
                        filter(models.AccountType.parent_id.isnot(None)).
                        order_by(models.AccountType.name).all()):
        choices.extend([(accounttype.id, accounttype.name)])

    return choices


def get_bankaccounttype_as_choices(first_choice=None):
    """Get bankaccounttypes, sort and return as form choices field."""

    choices = [(u'0', (first_choice or ''))]
    for bankaccounttype in models.db.session \
            .query(models.BankAccountType).all():
        choices.extend([(bankaccounttype.id, bankaccounttype.desc)])

    return choices


def get_bankaccount_as_choices():
    """Get bankaccount sort and return as form choices field."""
    choices = [(0, '')]
    for bankaccount in models.db.session.query(models.BankAccount) \
            .filter_by(user_id=g.user.id) \
            .order_by(models.BankAccount.bank,
                      models.BankAccount.number).all():
        choices.extend([(bankaccount.id,
                         '{0.bank} - {0.number}'.format(bankaccount))])
    return choices


def get_entities_as_choices():
    """Get entities sort and return as form choices field."""
    choices = [(0, '')]
    for entity in models.db.session.query(models.Entity) \
            .filter_by(user_id=g.user.id) \
            .order_by(models.Entity.name).all():
        choices.extend([(entity.id, entity.name)])
    return choices


def get_entitytype_as_choices(first_choice=None):
    """Get entitytypes sort and return as form choices field."""

    choices = [(u'0', (first_choice or ''))]
    for entitytype in models.db.session \
            .query(models.EntityType).all():
        choices.extend([(entitytype.id, entitytype.name)])

    return choices


def get_account_as_choices():
    """Get accounts sort and return as form choices field."""
    choices = [(0, '')]
    for account in models.db.session.query(models.Account) \
            .join(models.Entity,
                  models.Account.entity_id == models.Entity.id)  \
            .filter_by(user_id=g.user.id) \
            .order_by(models.Entity.name, models.Account.name).all():
        choices.extend([(account.id,
                         "{0.entity.name} - {0.name}".format(account))])
    return choices
