import pytest
from penny import create_app, user_datastore
from penny.models import db
from flask_migrate import upgrade
from flask import Flask
from flask_security.utils import hash_password
import tempfile
from penny.common.init_data import import_all_types
import os


@pytest.fixture
def app():

    db_fd, db_path = tempfile.mkstemp()

    app: Flask = create_app(
        dict(
            SQLALCHEMY_DATABASE_URI='sqlite:///{0}'.format(db_path),
            SQLALCHEMY_TRACK_MODIFICATIONS=False,
            SECRET_KEY="tatokuddMiWradfo",
            SECURITY_PASSWORD_SALT="cecsebWorPenitdiTin",
            DEBUG=True,
            TESTING=True,
            CSRF_ENABLED=False,
            WTF_CSRF_ENABLED=False,
            SECURITY_PASSWORD_HASH="plaintext",
            SECURITY_EMAIL_VALIDATOR_ARGS={"check_deliverability": False},
        )
    )


    with app.app_context():
        upgrade()
        user_datastore.create_user(email="test@example.org", password="secret")
        db.session.commit()
        import_all_types()

    yield app

    os.close(db_fd)
    os.unlink(db_path)


@pytest.fixture
def client(app):
    return app.test_client()


@pytest.fixture
def runner(app):
    return app.test_cli_runner()


class AuthActions:
    def __init__(self, client):
        self._client = client

    def login(self, username="test@example.org", password="secret"):
        data: dict[str, str] = {"email": username, "password": password}
        return self._client.post(
            '/login',
            data=data,
            follow_redirects=True
        )


@pytest.fixture
def auth(client):
    return AuthActions(client)


def create_entity(client):
    return client.post(
        "/entities/add",
        data={
            "name": "test-entity-name",
            "entitytype": 1
        },
        follow_redirects=True
    )


def create_bankaccount(client):
    return client.post(
        "/bankaccounts/add",
        data={
            "bank": "test-bank",
            "number": "test-bank-number",
            "desc": "test-bank-desc",
            "bankaccounttype": 1,
            "entitytype": 1,
            "total_balance": 0,
        },
        follow_redirects=True
    )


def create_account(client):
    return client.post(
        "/accounts/add",
        data={
            "name": "test-account-name",
            "desc": "test-account-desc",
            "accounttype": 2,
            "entity": 1,
        },
        follow_redirects=True
    )


def create_transaction(client, **kwargs):
    date = kwargs.get('date', '01/01/2020')
    memo = kwargs.get('memo', 'test-transaction-memo')
    credit = kwargs.get('credit', 0)
    debit = kwargs.get('debit', 0)
    account = kwargs.get('account', 0)
    bankaccount = kwargs.get('bankaccount', 0)
    follow_redirects = kwargs.get('follow_redirects', True)

    return client.post(
        "/transactions/add",
        data={
            "debit": debit,
            "credit": credit,
            "memo": "test-transaction-memo",
            "account": account,
            "bankaccount": bankaccount
        },
        follow_redirects=follow_redirects
    )
