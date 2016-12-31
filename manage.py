from flask.ext.script import Manager, prompt_bool
from flask.ext.migrate import Migrate, MigrateCommand

from app import app
from app.models import db

migrate = Migrate(app, db)
manager = Manager(app, db)
manager.add_command('db', MigrateCommand)


@manager.command
def dropdata():
    "drop all database tables."
    if prompt_bool("Are you sure you want to lose all your data?"):
        db.drop_all()
        db.engine.execute("DROP TABLE IF EXISTS alembic_version")

if __name__ == "__main__":
    manager.run()
