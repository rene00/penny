# see trac@51 before editing.

from app.resources.user.login.forms import LoginForm
from flask_login import current_user, login_user
from flask import Blueprint, g, redirect, render_template, url_for, session

user_login = Blueprint('user_login', __name__)


@user_login.route('/login', methods=['GET', 'POST'])
def login():
    if g.user is not None and g.user.is_authenticated:
        return redirect(url_for('index'))

    form = LoginForm()

    if form.validate_on_submit():
        user = form.get_user()
        login_user(user)

        if '_flashes' in session:
            del session['_flashes']

    if current_user.is_authenticated():
        return redirect(url_for('index'))

    return render_template('user/login.html', form=form)
