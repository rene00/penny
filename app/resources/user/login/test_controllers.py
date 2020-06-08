import pytest
from app.test_init import client


def test_login(client):
    rv = client.post(
        '/login',
        data=dict(
            email='test@example.org',
            password='secret'
        ),
        follow_redirects=True
    )
    assert b'Logout' in rv.data
