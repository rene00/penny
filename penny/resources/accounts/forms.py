from flask_wtf import FlaskForm
from werkzeug.datastructures import MultiDict
from wtforms import StringField, SelectField, validators
from uuid import uuid4


class FormAccount(FlaskForm):
    name = StringField(
        u"Account Name", default="", validators=[validators.DataRequired()]
    )
    desc = StringField(u"Account Description", default="", validators=[])
    accounttype = SelectField(
        u"Account Type", validators=[validators.DataRequired()], coerce=int
    )
    entity = SelectField(
        u"Account Entity Owner", validators=[validators.DataRequired()], coerce=int
    )

    def reset(self):
        blankdata = MultiDict([])
        self.process(blankdata)

    def set_defaults(self, account):
        if account.accounttype:
            self.accounttype.default = account.accounttype.id
        if account.entity:
            self.entity.default = account.entity.id
