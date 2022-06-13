from flask_wtf import FlaskForm
from wtforms import StringField, validators


class FormTag(FlaskForm):
    name = StringField("Tag Name", default="", validators=[validators.DataRequired()])
    desc = StringField("Tag Description", default="")
