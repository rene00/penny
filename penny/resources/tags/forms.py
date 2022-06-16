from flask_wtf import FlaskForm
from wtforms import StringField, validators
from werkzeug.datastructures import MultiDict


class FormTag(FlaskForm):
    name = StringField("Tag Name", default="", validators=[validators.DataRequired()])
    desc = StringField("Tag Description", default="")
    regex = StringField("Filter", default="")

    def reset(self):
        self.process(MultiDict([]))
