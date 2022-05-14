from flask_wtf import FlaskForm
from wtforms import TextAreaField, PasswordField, validators
from werkzeug.security import check_password_hash
from penny import models
from sqlalchemy.orm.exc import NoResultFound


class LoginForm(FlaskForm):
    email = TextAreaField(u"Email", default="", validators=[validators.DataRequired()])
    password = PasswordField(
        u"Password", default="", validators=[validators.DataRequired()]
    )

    def validate_email(self, field):
        user = self.get_user()

        if user is None:
            # invalid user
            raise validators.ValidationError("Incorrect username or password.")

        if not check_password_hash(user.password, self.password.data):
            # incorrect password
            raise validators.ValidationError("Incorrect username or password.")

        return user

    def get_user(self):
        try:
            user = models.db.session.query(models.User).filter_by(email=self.email.data)
        except NoResultFound:
            return None
        else:
            return user
