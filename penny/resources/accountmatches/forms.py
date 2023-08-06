from flask_wtf import FlaskForm
from werkzeug.datastructures import MultiDict
from wtforms import StringField, StringField, SelectField, validators


class FormAccountMatch(FlaskForm):
    name = StringField("Name", default="", validators=[validators.DataRequired()])
    desc = StringField("Description", default="", validators=[])
    account = SelectField("Account", validators=[validators.DataRequired()], coerce=int)
    bankaccount = SelectField(
        "Account Entity Owner", validators=[validators.DataRequired()], coerce=int
    )

    def reset(self):
        blankdata = MultiDict([])
        self.process(blankdata)

    def set_defaults(self, accountmatch):
        if accountmatch.account:
            self.account.default = accountmatch.account.id
        if accountmatch.bankaccount:
            self.bankaccount.default = accountmatch.bankaccount.id
        return self


class FormAccountMatchFilter(FlaskForm):
    regex = StringField("Filter", validators=[validators.DataRequired()])

    def reset(self):
        blankdata = MultiDict([])
        self.process(blankdata)
