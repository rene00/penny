import pytest
from flask import g, session


def test_login(client, auth):
    assert client.get("/login").status_code == 200

    response = auth.login()
    assert b'Logout' in response.data

    with client:
        client.get("/")
        assert g.user.id == 1
        assert g.user.email == "test@example.org"
