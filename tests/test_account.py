import pytest
from flask import json
from datetime import datetime
from .conftest import (
    create_entity,
    create_bankaccount,
    create_account
)

def test_account(client, auth):
    with client:
        auth.login()
        create_entity(client)
        create_bankaccount(client)
        response = create_account(client)
        assert b'value="test-account-name"' in response.data
        assert b'value="test-account-desc"' in response.data

        response = client.get("/data/accounts/?search=&order=desc&offset=0&limit=1")
        data = json.loads(response.data)
        assert 200 == response.status_code
        assert data['rows'][0]['name'] == 'test-account-name'
        assert data['rows'][0]['id'] == 1
        assert data['rows'][0]['entity_name'] == 'test-entity-name'
