import pytest
from penny.common import util
from penny import models
import datetime

@pytest.mark.parametrize("transaction, hash", [
    (
        models.Transaction(
            date=datetime.datetime.strptime('20200101', '%Y%m%d'),
            debit=-1600,
            credit=0,
            memo="stan.com.au Sydney AUS",
            bankaccount=models.BankAccount(id=1),
            fitid=None,
        ),
        "0928895e5604901c7e3ac6969d8bc66576c1479f7f5fdb38e5a401d33287e47a"
    ),
    (
        models.Transaction(
            date=datetime.datetime.strptime('20200101', '%Y%m%d'),
            debit=-1600,
            credit=0,
            memo="stan.com.au   Sydney AUS",
            bankaccount=models.BankAccount(id=1),
            fitid=None,
        ),
        "0928895e5604901c7e3ac6969d8bc66576c1479f7f5fdb38e5a401d33287e47a"
    ),
    (
        models.Transaction(
            date=datetime.datetime.strptime('20200101', '%Y%m%d'),
            debit=-1600,
            credit=0,
            memo="stan.com.au   Sydney AUS       ",
            bankaccount=models.BankAccount(id=1),
            fitid=None,
        ),
        "0928895e5604901c7e3ac6969d8bc66576c1479f7f5fdb38e5a401d33287e47a"
    ),
])
def test_generate_transaction_hash(transaction: models.Transaction, hash: str) -> None:
    assert util.generate_transaction_hash(transaction) == hash

def test_merge_dicts() -> None:
    assert util.merge_dicts({'a': 'b'}, {'c': 'd'}) == {'a': 'b', 'c': 'd'}
