from flask_security.datastore import SQLAlchemyUserDatastore
from penny import models
from penny.models import db

user_datastore = SQLAlchemyUserDatastore(db, models.User, models.Role)
