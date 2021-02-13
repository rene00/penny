from flask import Flask, g, request, redirect, url_for
from flask_mail import Mail
from flask_login import LoginManager, current_user
from flask_security import Security, SQLAlchemyUserDatastore
from flask_migrate import Migrate
from alembic import command
from penny import models, resources, util
from penny.models import db
from penny.lib.flask_security.register import ExtendedRegisterForm
from penny.common.init_data import import_all_types
import os
from pathlib import Path
import locale

PENNY_CONF_FILE = Path('/etc/penny/penny.conf.py')

migrate = Migrate()

user_datastore = SQLAlchemyUserDatastore(db, models.User, models.Role)


def create_app(test_config=None, skip_migrations=False):
    app = Flask(__name__)

    app.config.update(
        CSRF_ENABLED=True,
        WTF_CSRF_ENABLED=True,
        SECRET_KEY='s3cr3tk3y',
        SQLALCHEMY_DATABASE_URI='sqlite:///{0}'.format(
            os.path.join(
                os.path.abspath(os.path.dirname(__file__)),
                'penny.db'
            )
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
        DEBUG=True
    )

    if test_config is None:
        if PENNY_CONF_FILE.is_file():
            app.config.from_pyfile(PENNY_CONF_FILE)

        if os.environ.get('CONFIG_FILE'):
            app.config.from_envvar('CONFIG_FILE')
    else:
        app.config.from_mapping(test_config)

    db.init_app(app)
    migrate.init_app(app, db)

    if not skip_migrations:
        with app.app_context():
            config = migrate.get_config(None)
            command.upgrade(config, 'head', sql=False, tag=None)
            import_all_types()

    login_manager = LoginManager()
    login_manager.init_app(app)
    login_manager.login_message_category = 'debug'

    security = Security(
        app, user_datastore, register_form=ExtendedRegisterForm,
        confirm_register_form=ExtendedRegisterForm
    )

    @app.template_filter('convert_to_float')
    def convert_to_float(cents):
        return util.convert_to_float(cents)

    @app.template_filter('convert_to_float_positive')
    def convert_to_float_positive(cents):
        locale.setlocale(locale.LC_ALL, 'en_AU.UTF-8')
        return locale.currency(abs(float(cents / float(100))), grouping=True)

    @login_manager.user_loader
    def load_user(id):
        return models.User.get(id)

    @app.before_request
    def load_current_user():
        g.user = current_user

        if not current_user.is_authenticated:
            return None

        # Ensure user is logged in at all times.
        allowed_urls = ['robots.txt', 'security.login', 'static',
                        'security.register', 'index']

        if request.endpoint not in allowed_urls and \
                not current_user.is_authenticated:
            app.logger.info("Anonymous user attempted to access secured "
                            "route: {0}".format(request.endpoint))
            return redirect(url_for('user_login.login'))

        # Check if user is enabled for alpha. If they are not enabled, log
        # them out and redirect them to login.
        if not g.user.alpha_enabled and request.endpoint not in allowed_urls:
            app.logger.info("User not enabled for alpha: id={0},{1}".
                            format(g.user.id, request.endpoint))
            logout_user()
            return redirect(url_for('user_login.login'))

    @app.route('/', methods=['GET'])
    def index():
        if not current_user.is_authenticated:
            return redirect(url_for('user_login.login'))
        return redirect(url_for('bankaccounts._bankaccounts'))

    @app.route('/robots.txt')
    def static_from_root():
        return send_from_directory(app.static_folder, request.path[1:])

    @app.route('/post_confirm_view')
    def post_confirm_view():
        """Log user out once they confirm.

        This will force the user to log in with their credentials once
        they have confirmed.
        """
        if g.user:
            logout_user()
        return redirect(url_for('user_login.login'))

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
