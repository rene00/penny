from typing import Optional
import requests
from sqlalchemy import select
from sqlalchemy.exc import IntegrityError
from sqlalchemy.orm.exc import NoResultFound
from penny import models


class TXResponse:
    def __init__(self, body: dict):
        self.body = body

    @property
    def sa4_name(self) -> Optional[str]:
        return self.locality.get("sa4", {}).get("name", None)

    @property
    def sa3_name(self) -> Optional[str]:
        return self.locality.get("sa3", {}).get("name", None)

    @property
    def locality(self) -> dict:
        return self.body.get("locality", {}).get("names", {}).get("100", {})

    @property
    def postcode(self) -> Optional[int]:
        return self.locality.get("postcode", None)

    @property
    def state(self) -> Optional[str]:
        return self.body.get("locality", {}).get("state", {}).get("name", None)


class TXMeta:
    def __init__(self, url: str):
        self.url = url
        self.response: dict
        self.txresponse: TXResponse

    def request(self, memo: str) -> None:
        try:
            response = requests.post(self.url, json={"memo": memo})
            response.raise_for_status()
        except (
            requests.exceptions.HTTPError,
            requests.exceptions.RequestException,
        ) as e:
            raise e

        self.response = response.json()
        self.txresponse = TXResponse(response.json())

    def persist(self, transaction, session):
        for k, v in {
            "postcode": self.txresponse.postcode,
            "state": self.txresponse.state,
            "sa3_name": self.txresponse.sa3_name,
            "sa4_name": self.txresponse.sa4_name,
        }.items():
            if v is None:
                continue

            try:
                rows = session.execute(
                    select(models.TransactionMetaType).where(
                        models.TransactionMetaType.name == k
                    )
                ).one()
            except NoResultFound as e:
                raise e
            transaction_meta_type = rows[0]

            session.add(
                models.TransactionMeta(
                    tx_id=transaction.id,
                    tx_meta_type_id=transaction_meta_type.id,
                    value=v,
                )
            )

            try:
                session.commit()
            except IntegrityError:
                session.rollback()
                continue
