from flask_wtf import Form
from werkzeug.datastructures import MultiDict
from wtforms import StringField, TextField, SelectField, validators


class FormAccountMatch(Form):
    name = TextField(u'Name', default='', validators=[validators.Required()])
    desc = TextField(u'Description', default='', validators=[])
    account = SelectField(u'Account', validators=[validators.Required()],
                          coerce=int)
    bankaccount = SelectField(u'Account Entity Owner',
                              validators=[validators.Required()], coerce=int)

    def reset(self):
        blankdata = MultiDict([])
        self.process(blankdata)

    def set_defaults(self, accountmatch):
        if accountmatch.account:
            self.account.default = accountmatch.account.id
        if accountmatch.bankaccount:
            self.bankaccount.default = accountmatch.bankaccount.id
        return self


class FormAccountMatchFilter(Form):
    regex = StringField(u'Filter', validators=[validators.Required()])

    def reset(self):
        blankdata = MultiDict([])
        self.process(blankdata)
