# Functions that are responsible for populating default types within the
# database.
from flask import current_app as app
from flask_sqlalchemy import SQLAlchemy
from penny import models
from sqlalchemy import select
from sqlalchemy.exc import IntegrityError
from sqlalchemy.orm.exc import NoResultFound
from typing import Tuple


def import_accounttypes(db):
    """Import default accounttypes."""
    accounttypes = {
        "Liabilities": ("Current Liability", "Liability", "Non-current Liability"),
        "Assets": ("Current Asset", "Fixed Asset", "Non-current Asset", "Prepayment"),
        "Revenue": ("Other Income", "Revenue", "Sale"),
        "Transfer": ("Bank Transfer",),
        "Expenses": ("Depreciation", "Direct Costs", "Expense", "Overhead"),
        "Equity": ("Equity", "Stocks"),
    }

    for parent, children in accounttypes.items():
        try:
            _parent = (
                db.session.query(models.AccountType)
                .filter_by(name=parent, parent_id=None)
                .one()
            )
        except NoResultFound:
            _parent = models.AccountType(name=parent, parent_id=None)
            db.session.add(_parent)

        for child in children:
            try:
                _child = (
                    db.session.query(models.AccountType)
                    .filter_by(name=child, parent=_parent)
                    .one()
                )
            except NoResultFound:
                _child = models.AccountType(name=child, parent=_parent)
                db.session.add(_child)

        try:
            db.session.commit()
        except IntegrityError:
            db.session.rollback()
            app.logger.error("Failed to import accounttypes.")


def import_bankaccounttypes(db):
    """Import bankaccounttypes."""
    bankaccounttypes = {
        "Savings": "Savings",
        "Credit Loan": "Loan",
        "Credit Card": "Credit Card",
        "Cash": "Cash",
    }

    for name, desc in bankaccounttypes.items():
        try:
            (db.session.query(models.BankAccountType).filter_by(name=name).one())
        except NoResultFound:
            bankaccounttype = models.BankAccountType(name=name, desc=desc)
            db.session.add(bankaccounttype)
            try:
                db.session.commit()
            except IntegrityError:
                db.session.rollback()
                app.logger.error("Failed to import bankaccounttypes.")


def import_entitytypes(db):
    entitytypes = ("Company", "Person", "Sole Trader")

    for name in entitytypes:
        try:
            (db.session.query(models.EntityType).filter_by(name=name).one())
        except NoResultFound:
            entitytype = models.EntityType(name=name)
            db.session.add(entitytype)
            try:
                db.session.commit()
            except IntegrityError:
                db.session.rollback()
                app.logger.error("Failed to import entitytypes.")


def import_tx_meta_name(db: SQLAlchemy) -> None:
    """Import Transaction Meta types"""
    for name in ("locality_name", "postcode", "state", "sa3_name", "sa4_name"):
        try:
            db.session.execute(
                select(models.TransactionMetaType).where(
                    models.TransactionMetaType.name == name
                )
            ).one()
        except NoResultFound:
            db.session.add(models.TransactionMetaType(name=name))
            try:
                db.session.commit()
            except IntegrityError:
                db.session.rollback()
                app.logger.error("Failed to import transaction meta name.")


def import_all_types() -> None:
    import_accounttypes(models.db)
    import_bankaccounttypes(models.db)
    import_entitytypes(models.db)
    import_tx_meta_name(models.db)
