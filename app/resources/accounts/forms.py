from flask_wtf import Form
from werkzeug.datastructures import MultiDict
from wtforms import TextField, SelectField, HiddenField, validators
from uuid import uuid4


class FormAccount(Form):
    code = HiddenField()
    name = TextField(u'Account Name', default='',
                     validators=[validators.Required()])
    desc = TextField(u'Account Description', default='', validators=[])
    accounttype = SelectField(u'Account Type',
                              validators=[validators.Required()], coerce=int)
    entity = SelectField(u'Account Entity Owner',
                         validators=[validators.Required()], coerce=int)

    def reset(self):
        blankdata = MultiDict([])
        self.process(blankdata)

    def set_defaults(self, account):
        if account.accounttype:
            self.accounttype.default = account.accounttype.id
        if account.entity:
            self.entity.default = account.entity.id

    # trac@27: ive refactored out the account code. in the future when i
    # have customers the account code may become important though for
    # now it will be just TBD.
    def get_tbd_code(self):
        return '{uuid}'.format(uuid=uuid4())
