from sqlalchemy.orm.exc import NoResultFound
from flask import Flask
import requests
from requests import Response
import sys
from penny.models import db
from typing import Optional, Dict
import penny
from penny import models


def fetch_tx_meta(tx_id: int) -> None:
    app: Flask = penny.create_app()
    metaurl: str = app.config.get("TX_META_TXACCT_URL", None)
    if metaurl is None:
        print("flask config TX_META_TXACCT_URL not set", file=sys.stderr)
        return

    with app.app_context():
        transaction: models.Transaction
        try:
            transaction = models.Transaction.query.filter_by(id=tx_id).one()
        except NoResultFound:
            print(f"transaction not found: id={tx_id}", file=sys.stderr)
            return

        try:
            response: Response = requests.post(metaurl, json={"memo": transaction.memo})
            response.raise_for_status()
        except (requests.exceptions.HTTPError, requests.exceptions.RequestException) as e:
            print(e, file=sys.stderr)
            return

        data: Dict = response.json()
        state_name: Optional[Dict] = data.get("locality", {}).get("state", {}).get("name", None)

        for weight, locality_name in data.get("locality", {}).get("names", {}).items():
            if weight != "100":
                continue

            """
            postcode: Optional[models.TxMetaPostCode] = models.TxMetaPostCode.query.filter_by(name=locality_name.get("name"), postcode=locality_name.get("postcode")).one_or_none()
            if postcode is None:
                postcode: Optional[models.TxMetaPostCode] = models.TxMetaPostCode(
                    name=locality_name.get("name"),
                    postcode=locality_name.get("postcode"),
                )


            sa3: Optional[models.TxMetaSa3] = None
            if "sa3" in locality_name:
                sa3 = models.TxMetaSa3.query.filter_by(name=locality_name["sa3"].get("name")).one_or_none()
                if sa3 is None:
                    sa3 = models.TxMetaSa3(name=locality_name["sa3"].get("name"))
                    db.session.add(sa3)

            state: Optional[models.TxMetaState] = None
            if state_name is not None:
                state: Optional[models.TxMetaState] = models.TxMetaState.query.filter_by(name=state_name).
                db.session.add(state)

            locality: models.TxMetaLocality = models.TxMetaLocality(weight=weight, transaction=transaction, postcode=postcode, sa3=sa3, state=state)
            db.session.add(locality)
            try:
                db.session.commit()
            except IntegrityError as e:
                db.session.rollback()
                print(f"failed to store locality: {e}", file=sys.stderr)
            """
