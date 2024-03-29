from flask import Flask, g, redirect, url_for

from flask_login import current_user
from flask_security.core import Security
from flask_security.datastore import SQLAlchemyUserDatastore
from flask_security.decorators import auth_required
from flask_migrate import Migrate
from penny import resources, util
from penny.extensions import user_datastore
from penny.models import db
from penny.common.init_data import import_all_types
from penny.cli.task import task_cli
from penny.cli.seed import seed_cli
from penny.cli.report import report_cli
from penny.cli.txmeta import txmeta_cli
from penny.cli.transaction import transaction_cli
import os
from pathlib import Path
import locale

PENNY_CONF_FILE = Path("/etc/penny/penny.conf.py")

migrate = Migrate()


def create_app(test_config=None):
    app = Flask(__name__)

    if test_config is None:
        if PENNY_CONF_FILE.is_file():
            app.config.from_pyfile(PENNY_CONF_FILE)

        if os.environ.get("CONFIG_FILE"):
            app.config.from_envvar("CONFIG_FILE")
    else:
        app.config.from_mapping(test_config)

    app.cli.add_command(task_cli)
    app.cli.add_command(seed_cli)
    app.cli.add_command(report_cli)
    app.cli.add_command(txmeta_cli)
    app.cli.add_command(transaction_cli)

    db.init_app(app)
    migrate.init_app(app, db)
    security = Security(app, user_datastore)

    if os.environ.get("PENNY_IMPORT_ALL_TYPES"):
        with app.app_context():
            import_all_types()

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

    app.register_blueprint(resources.accounts)
    app.register_blueprint(resources.accountmatches)
    app.register_blueprint(resources.bankaccounts)
    app.register_blueprint(resources.data_accounts)
    app.register_blueprint(resources.data_accountmatches)
    app.register_blueprint(resources.data_bankaccounts)
    app.register_blueprint(resources.data_entities)
    app.register_blueprint(resources.data_reports)
    app.register_blueprint(resources.data_tags)
    app.register_blueprint(resources.data_transactions)
    app.register_blueprint(resources.entities)
    app.register_blueprint(resources.reports)
    app.register_blueprint(resources.tags)
    app.register_blueprint(resources.tasks)
    app.register_blueprint(resources.transactions)
    app.register_blueprint(resources.user_login)
    app.register_blueprint(resources.user_logout)

    return app
