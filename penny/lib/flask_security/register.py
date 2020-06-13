from flask_security.forms import RegisterForm
from wtforms import StringField, validators


class ExtendedRegisterForm(RegisterForm):
    """ Extend flask-security registery field.

    From https://pythonhosted.org/Flask-Security/customizing.html

    """
    first_name = StringField('First Name', [validators.DataRequired()])
    last_name = StringField('Last Name', [validators.DataRequired()])
