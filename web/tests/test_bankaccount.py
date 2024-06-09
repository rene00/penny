import pytest
from flask import g, session
from datetime import datetime
from .conftest import create_entity, create_bankaccount

def test_bankaccount(client, auth):
    auth.login()
    create_entity(client)
    response = create_bankaccount(client)
    assert b'value="test-bank"' in response.data
    assert b'value="test-bank-desc"' in response.data
    assert b'value="test-bank-number"' in response.data
