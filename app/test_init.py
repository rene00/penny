import pytest
from app import create_app, user_datastore
from app.models import db
import os
import tempfile


@pytest.fixture
def client():

    db_fd, db_path = tempfile.mkstemp()

    app = create_app(
        dict(
            SQLALCHEMY_DATABASE_URI='sqlite:///{0}'.format(db_path),
            CSRF_ENABLED=False,
            WTF_CSRF_ENABLED=False,
        )
    )

    with app.app_context():
        user_datastore.create_user(email="test@example.org", password="secret")
        db.session.commit()

    with app.test_client() as client:
        yield client

    os.close(db_fd)
    os.unlink(db_path)
