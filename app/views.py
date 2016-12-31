from app import app, login_manager, models
from flask import g, redirect, request, send_from_directory, url_for
from flask_login import current_user
from flask_security import logout_user
import locale


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


@login_manager.user_loader
def load_user(id):
    return models.User.get(id)


@app.route('/', methods=['GET'])
def index():
    if not current_user.is_authenticated:
        return redirect(url_for('user_login.login'))
    return redirect(url_for('bankaccounts._bankaccounts'))


@app.route('/robots.txt')
def static_from_root():
    return send_from_directory(app.static_folder, request.path[1:])


@app.template_filter('convert_to_float')
def convert_to_float(cents):
    locale.setlocale(locale.LC_ALL, 'en_AU.UTF-8')
    return locale.currency(float(cents / float(100)), grouping=True)


@app.template_filter('convert_to_float_positive')
def convert_to_float_positive(cents):
    locale.setlocale(locale.LC_ALL, 'en_AU.UTF-8')
    return locale.currency(abs(float(cents / float(100))), grouping=True)


@app.route('/post_confirm_view')
def post_confirm_view():
    """Log user out once they confirm.

    This will force the user to log in with their credentials once
    they have confirmed.
    """
    if g.user:
        logout_user()
    return redirect(url_for('user_login.login'))
