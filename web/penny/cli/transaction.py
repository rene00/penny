from flask.cli import AppGroup
from penny import models
from sqlalchemy import select
import penny
import click

transaction_cli = AppGroup("transaction")


@transaction_cli.command("get")
@click.option("--id", type=int)
def transaction_get(id: int) -> None:
    app = penny.create_app()
    with app.app_context():
        row = models.db.session.execute(
            select(models.Transaction).where(models.Transaction.id == id)
        ).one()
        transaction = row[0]

        print(transaction)
        print(transaction.bankaccount)
        print(transaction.account)
        print(transaction.notes)
        print(transaction.meta)
