import pytest
from flask import g, session
from datetime import datetime
from .conftest import create_account

def test_account(client, auth):
    auth.login()
    response = create_account(client)
    assert b'value="test-account-name"' in response.data
    assert b'value="test-account-desc"' in response.data
