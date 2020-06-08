from flask_wtf import Form
from werkzeug.datastructures import MultiDict
from wtforms import TextField, SelectField, validators


class FormEntity(Form):
    name = TextField(
        u'Name',
        default='',
        validators=[validators.DataRequired()]
    )
    entitytype = (SelectField(u'Type', validators=[], coerce=int))

    def reset(self):
        blankdata = MultiDict([])
        self.process(blankdata)

    def set_defaults(self, entity):
        if entity.entitytype:
            self.entitytype.default = entity.entitytype.id
