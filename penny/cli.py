from penny import models
from flask.cli import AppGroup
import random


seed_cli = AppGroup("seed")


@seed_cli.command("init")
def seed_init() -> None:
    user: models.User = models.db.session.query(models.User).filter_by(id=1).one()

    entity_type: models.EntityType = (
        models.db.session.query(models.EntityType).filter_by(name="Person").one()
    )

    entity: models.Entity = models.Entity.query.filter_by(
        name="test person entity", user_id=user.id
    ).first()
    if entity is None:
        entity = models.Entity(
            name="test person entity", entitytype_id=entity_type.id, user_id=user.id
        )
    models.db.session.add(entity)

    bankaccount_type: models.BankAccountType = (
        models.db.session.query(models.BankAccountType)
        .filter_by(name="Credit Card")
        .one()
    )

    bankaccounts: list[models.BankAccount] = [
        models.BankAccount(
            user_id=user.id,
            number="test number 1",
            bank="test bank 1",
            bankaccounttype_id=bankaccount_type.id,
            entity_id=entity.id,
            desc="test desc 1",
        )
    ]

    for i in bankaccounts:
        bankaccount: models.BankAccount = models.BankAccount.query.filter_by(
            user_id=i.user_id,
            number=i.number,
            bank=i.bank,
            bankaccounttype_id=i.bankaccounttype_id,
            entity_id=i.entity_id,
            desc=i.desc,
        ).first()
        if bankaccount is None:
            models.db.session.add(i)

    account_types: list[models.AccountType] = models.AccountType.query.filter_by(
        name="Expense"
    ).all()

    accounts: list[models.Account] = [
        models.Account(
            user_id=user.id,
            name="test name 1",
            desc="test desc 1",
            accounttype_id=random.choice(account_types).id,
            entity_id=entity.id,
        )
    ]

    for i in accounts:
        models.db.session.add(i)
        models.db.session.commit()

    transactions: list[models.Transaction] = [
        models.Transaction(
            user_id=user.id,
            debit=-100,
            credit=0,
            memo="test memo 1",
            bankaccount_id=random.choice(bankaccounts).id,
            account_id=random.choice(accounts).id,
        )
    ]

    for i in transactions:
        models.db.session.add(i)
        models.db.session.commit()

    models.db.session.commit()

    return None