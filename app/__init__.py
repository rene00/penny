from flask import Flask
from flask_mail import Mail
from flask_login import LoginManager
from flask_security import Security, SQLAlchemyUserDatastore
from flask_migrate import Migrate
from alembic import command
from app import models, resources
from app.models import db
from app.lib.flask_security.register import ExtendedRegisterForm
from app.common.init_data import import_all_types
import os
from pathlib import Path

PENNY_CONF_FILE = Path('/etc/penny/penny.conf.py')

app = Flask(__name__)

app.config.update(
    CSRF_ENABLED=True,
    SECRET_KEY='s3cr3tk3y',
    SQLALCHEMY_DATABASE_URI='sqlite:///{0}'.format(
        os.path.join(os.path.abspath(os.path.dirname(__file__)), 'penny.db')
    ),
    SQLALCHEMY_TRACK_MODIFICATIONS=False,
    SECURITY_CONFIRMABLE=False,
    SECURITY_LOGIN_USER_TEMPLATE='user/login.html',
    SECURITY_PASSWORD_HASH='bcrypt',
    SECURITY_PASSWORD_SALT='s4lt',
    SECURITY_POST_CONFIRM_VIEW='post_confirm_view',
    SECURITY_REGISTERABLE=True,
    SECURITY_REGISTER_USER_TEMPLATE='user/register.html',
    SECURITY_SEND_PASSWORD_CHANGE_EMAIL=False,
    SECURITY_SEND_PASSWORD_RESET_NOTICE_EMAIL=False,
    SECURITY_SEND_REGISTER_EMAIL=False,
    TRANSACTION_ATTACHMENTS_UPLOAD_FOLDER='files/attachments',
    TRANSACTION_UPLOADS_UPLOAD_FOLDER='files/uploads',
    RQ_DEFAULT_URL='redis://localhost:6379/0',
    DEBUG = False
)

if PENNY_CONF_FILE.is_file():
    app.config.from_pyfile(PENNY_CONF_FILE)

if os.environ.get('CONFIG_FILE'):
    app.config.from_envvar('CONFIG_FILE')

db.init_app(app)
migrate = Migrate(app, db)
with app.app_context():
    config = migrate.get_config(None)
    command.upgrade(config, 'head', sql=False, tag=None)
    import_all_types()

mail = Mail(app)

login_manager = LoginManager()
login_manager.init_app(app)
login_manager.login_message_category = 'debug'

user_datastore = SQLAlchemyUserDatastore(db, models.User, models.Role)

security = Security(
    app, user_datastore, register_form=ExtendedRegisterForm,
    confirm_register_form=ExtendedRegisterForm
)

from app import views  # noqa

# resource blueprints.
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
