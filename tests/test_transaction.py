import pytest
from flask import g, session, json
from datetime import datetime
from .conftest import (
    create_account,
    create_bankaccount,
    create_entity,
    create_transaction
)


def test_transaction(client, auth):
    with client:
        auth.login()
        create_entity(client)
        create_bankaccount(client)
        create_account(client)
        response = create_transaction(
            client,
            memo="test-transaction-memo",
            credit=100
        )
        assert 200 == response.status_code

        response = client.get("/data/transactions/?search=&order=desc&offset=0&limit=1")
        data = json.loads(response.data)
        assert 200 == response.status_code

        tx = data['rows'][0]
        total = data['total']
        assert tx['memo'] == "test-transaction-memo"
        assert tx['memo_as_html'] == '<a href="/transactions/1">test-transaction-memo</a>'
        assert tx['id'] == 1
        assert tx['bankaccount'] is None
        assert tx['account'] is None
        assert tx['account_as_html'] == ''
        assert tx['credit'] == "$100.00"
        assert tx['fitid'] is None
        assert total == 1
