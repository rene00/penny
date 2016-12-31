from flask_wtf import Form
from werkzeug.datastructures import MultiDict
from wtforms import TextField, SelectField, validators


class FormBankAccount(Form):
    bank = TextField(u'Bank', default='', validators=[])
    number = TextField(u'Bank Account Number', default='',
                       validators=[validators.Required()])
    desc = TextField(u'Bank Account Description', default='', validators=[])

    bankaccounttype = (SelectField(u'Bank Account Type', validators=[],
                                   coerce=int))
    entity = (SelectField(u'Entity', validators=[validators.Required()],
                          coerce=int))
    total_balance = TextField(u'Total Balance')

    def reset(self):
        blankdata = MultiDict([])
        self.process(blankdata)

    def set_defaults(self, bankaccount):
        if bankaccount.bankaccounttype:
            self.bankaccounttype.default = bankaccount.bankaccounttype.id
        if bankaccount.entity:
            self.entity.default = bankaccount.entity.id
