from flask import Flask
from flask_mail import Mail
from flask_login import LoginManager
from flask_security import Security, SQLAlchemyUserDatastore
from flask_log import Logging
from app import models, resources
from app.models import db
from app.lib.flask_security.register import ExtendedRegisterForm


app = Flask(__name__)
app.config.from_envvar('CONFIG_FILE')

LOGGING_FMT = ("%(asctime)s - %(pathname)s:%(lineno)d "
               "- %(levelname)s: %(message)s")

flask_log = Logging(app)

db.init_app(app)

mail = Mail(app)

login_manager = LoginManager()
login_manager.init_app(app)
login_manager.login_message_category = 'debug'

user_datastore = SQLAlchemyUserDatastore(db, models.User, models.Role)

security = Security(
    app, user_datastore, register_form=ExtendedRegisterForm,
    confirm_register_form=ExtendedRegisterForm
)

from app import views

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
