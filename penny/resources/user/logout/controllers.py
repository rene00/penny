# see trac@51 before editing.

from flask_login import logout_user
from flask import Blueprint, redirect, url_for

user_logout = Blueprint('user_logout', __name__)


@user_logout.route('/logout')
def logout():
    logout_user()
    return redirect(url_for('user_login.login'))
