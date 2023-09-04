from sqlalchemy.orm.exc import NoResultFound
from flask import Flask
import sys
import penny
from penny.lib.txmeta import TXMeta
from sqlalchemy import select
from penny import models


def fetch_tx_meta(id: int) -> None:
    app: Flask = penny.create_app()
    url = app.config.get("TX_META_TXACCT_URL", None)
    if url is None:
        print("flask config TX_META_TXACCT_URL not set", file=sys.stderr)
        return

    with app.app_context():
        try:
            rows = models.db.session.execute(
                select(models.Transaction).where(
                    models.Transaction.id == id,
                    models.Transaction.is_deleted == False,
                    models.Transaction.is_archived == False,
                )
            ).one()
        except NoResultFound:
            app.logger.error(f"transaction not found: id={id}")
            return

        transaction = rows[0]

        txmeta = TXMeta(url)
        txmeta.request(transaction.memo)
        txmeta.persist(transaction, models.db.session)
