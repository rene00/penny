from flask.cli import AppGroup
from sqlalchemy.orm.exc import NoResultFound
from penny import models
from penny.lib.txmeta import TXMeta
import penny
import click
import json
from sqlalchemy import select

txmeta_cli = AppGroup("txmeta")


@txmeta_cli.command("transaction")
@click.option("--memo", type=str, default="")
@click.option("--id", type=int, default=0)
def txmeta_transaction(memo: str, id: int) -> None:
    """txmeta transaction posts the transaction to the txmeta url and prints the returned json to stdout"""
    app = penny.create_app()

    if memo != "" and id != 0:
        raise click.ClickException("cant set memo and id options")

    with app.app_context():
        url = app.config.get("TX_META_TXACCT_URL", None)
        if url is None:
            raise click.ClickException("TX_META_TXACCT_URL not set")

        m = memo
        if id != 0:
            try:
                rows = models.db.session.execute(
                    select(models.Transaction).where(models.Transaction.id == id)
                ).one()
            except NoResultFound as e:
                raise click.ClickException(str(e))
            m = rows[0].memo

        txmeta = TXMeta(url)
        txmeta.request(m)

        print(json.dumps(txmeta.response, indent=4))
        print(txmeta.txresponse.state)
        print(txmeta.txresponse.postcode)
        print(txmeta.txresponse.sa3_name)
        print(txmeta.txresponse.sa4_name)
        txmeta.persist(rows[0], models.db.session)
