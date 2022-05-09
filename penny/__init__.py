from flask import Flask, g, request, redirect, url_for

from flask_login import current_user
from flask_security.models import fsqla_v2 as fsqla
from flask_security import (
    Security,
    SQLAlchemyUserDatastore,
    auth_required,
    hash_password,
)
from flask_migrate import Migrate
from penny import models, resources, util
from penny import models
from penny.models import db
from penny.common.init_data import import_all_types
import os
from pathlib import Path
import locale

PENNY_CONF_FILE = Path("/etc/penny/penny.conf.py")

migrate = Migrate()

user_datastore = SQLAlchemyUserDatastore(db, models.User, models.Role)


def create_app(test_config=None, skip_migrations=False):
    app = Flask(__name__)

    if test_config is None:
        if PENNY_CONF_FILE.is_file():
            app.config.from_pyfile(PENNY_CONF_FILE)

        if os.environ.get("CONFIG_FILE"):
            app.config.from_envvar("CONFIG_FILE")
    else:
        app.config.from_mapping(test_config)

    db.init_app(app)
    migrate.init_app(app, db)
    security = Security(app, user_datastore)

    if os.environ.get("PENNY_IMPORT_ALL_TYPES"):
        with app.app_context():
            import_all_types()

    @app.before_first_request
    def create_user():
        db.create_all()
        if not user_datastore.find_user(email="rene@compounddata.com"):
            user_datastore.create_user(
                email="rene@compounddata.com", password=hash_password("lkjh0987")
            )
        db.session.commit()

    @app.before_request
    def load_current_user():
        g.user = current_user

    @app.template_filter("convert_to_float")
    def convert_to_float(cents):
        return util.convert_to_float(cents)

    @app.template_filter("convert_to_float_positive")
    def convert_to_float_positive(cents):
        locale.setlocale(locale.LC_ALL, "en_AU.UTF-8")
        return locale.currency(abs(float(cents / float(100))), grouping=True)

    @app.route("/", methods=["GET"])
    @auth_required()
    def index():
        return redirect(url_for("bankaccounts._bankaccounts"))

    @app.route("/post_confirm_view")
    def post_confirm_view():
        """Log user out once they confirm.

        This will force the user to log in with their credentials once
        they have confirmed.
        """
        if g.user:
            logout_user()
        return redirect(url_for("user_login.login"))

    app.register_blueprint(resources.accounts)
    app.register_blueprint(resources.accountmatches)
    app.register_blueprint(resources.bankaccounts)
    app.register_blueprint(resources.data_accounts)
    app.register_blueprint(resources.data_accountmatches)
    app.register_blueprint(resources.data_bankaccounts)
    app.register_blueprint(resources.data_entities)
    app.register_blueprint(resources.data_reports)
    app.register_blueprint(resources.data_transactions)
    app.register_blueprint(resources.entities)
    app.register_blueprint(resources.reports)
    app.register_blueprint(resources.tasks)
    app.register_blueprint(resources.transactions)
    app.register_blueprint(resources.user_login)
    app.register_blueprint(resources.user_logout)

    return app
