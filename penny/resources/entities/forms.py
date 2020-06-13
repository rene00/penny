from flask_wtf import FlaskForm
from werkzeug.datastructures import MultiDict
from wtforms import StringField, SelectField, validators


class FormEntity(FlaskForm):
    name = StringField(
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
