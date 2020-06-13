from flask_wtf import FlaskForm
from werkzeug.datastructures import MultiDict
from wtforms import StringField, SelectField, validators


class FormBankAccount(FlaskForm):
    bank = StringField(u'Bank', default='', validators=[])
    number = StringField(
        u'Bank Account Number',
        default='',
        validators=[validators.DataRequired()]
    )
    desc = StringField(u'Bank Account Description', default='', validators=[])

    bankaccounttype = (SelectField(u'Bank Account Type', validators=[],
                                   coerce=int))
    entity = (SelectField(u'Entity', validators=[validators.DataRequired()],
                          coerce=int))
    total_balance = StringField(u'Total Balance')

    def reset(self):
        blankdata = MultiDict([])
        self.process(blankdata)

    def set_defaults(self, bankaccount):
        if bankaccount.bankaccounttype:
            self.bankaccounttype.default = bankaccount.bankaccounttype.id
        if bankaccount.entity:
            self.entity.default = bankaccount.entity.id
