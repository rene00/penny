import pytest
from flask import g, session
from datetime import datetime
from .conftest import create_entity

def test_entity(client, auth):
    auth.login()
    response = create_entity(client)
    assert b'value="test-entity-name"' in response.data
