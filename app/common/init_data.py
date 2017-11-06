# Functions that are responsible for populating default types within the
# database.
from flask import current_app as app
from app import models
from sqlalchemy.exc import IntegrityError
from sqlalchemy.orm.exc import NoResultFound


def import_accounttypes():
    """Import default accounttypes."""
    accounttypes = {
        'Liabilities': ('Current Liability', 'Liability',
                        'Non-current Liability'),
        'Assets': ('Current Asset', 'Fixed Asset', 'Non-current Asset',
                   'Prepayment'),
        'Revenue': ('Other Income', 'Revenue', 'Sale'),
        'Transfer': ('Bank Transfer',),
        'Expenses': ('Depreciation', 'Direct Costs', 'Expense',
                     'Overhead'),
        'Equity': ('Equity', 'Stocks')
    }

    for parent, children in accounttypes.items():
        try:
            _parent = (
                models.db.session.query(models.AccountType).
                filter_by(name=parent, parent_id=None).one()
            )
        except NoResultFound:
            _parent = models.AccountType(name=parent, parent_id=None)
            models.db.session.add(_parent)

        for child in children:
            try:
                _child = (
                    models.db.session.query(models.AccountType).
                    filter_by(name=child, parent=_parent).one()
                )
            except NoResultFound:
                _child = models.AccountType(name=child, parent=_parent)
                models.db.session.add(_child)

        try:
            models.db.session.commit()
        except IntegrityError:
            models.db.session.rollback()
            app.logger.error("Failed to import accounttypes.")


def import_bankaccounttypes():
    """Import bankaccounttypes."""
    bankaccounttypes = {
        'Savings': 'Savings',
        'Credit Loan': 'Loan',
        'Credit Card': 'Credit Card',
        'Cash': 'Cash'
    }

    for name, desc in bankaccounttypes.items():
        try:
            (models.db.session.query(models.BankAccountType).
                filter_by(name=name).one())
        except NoResultFound:
            bankaccounttype = models.BankAccountType(name=name, desc=desc)
            models.db.session.add(bankaccounttype)
            try:
                models.db.session.commit()
            except IntegrityError:
                models.db.session.rollback()
                app.logger.error("Failed to import bankaccounttypes.")


def import_entitytypes():
    entitytypes = ('Company', 'Person', 'Sole Trader')

    for name in entitytypes:
        try:
            (models.db.session.query(models.EntityType).
             filter_by(name=name).one())
        except NoResultFound:
            entitytype = models.EntityType(name=name)
            models.db.session.add(entitytype)
            try:
                models.db.session.commit()
            except IntegrityError:
                models.db.session.rollback()
                app.logger.error("Failed to import entitytypes.")


def import_all_types():
    import_accounttypes()
    import_bankaccounttypes()
    import_entitytypes()
