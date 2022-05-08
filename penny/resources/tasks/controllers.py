from penny.common.tasks import run_accountmatchrun
from flask import Blueprint, g, flash, render_template, request, current_app as app
from flask_security.decorators import auth_required
from flask_wtf import FlaskForm
from werkzeug.datastructures import MultiDict
from wtforms import BooleanField
from rq import Queue
from redis import Redis


tasks = Blueprint("tasks", __name__)


class FormTasks(FlaskForm):
    accountmatch = BooleanField("Process Account Matches")

    def reset(self):
        self.process(MultiDict([]))


@tasks.route("/tasks", methods=["GET", "POST"])
@auth_required()
def _tasks():
    form = FormTasks()
    if form.validate_on_submit():
        if form.accountmatch.data:
            q = Queue(connection=Redis.from_url(app.config["REDIS_URL"]))
            task = q.enqueue(run_accountmatchrun, g.user.id)
            flash("Submitted Proccess Account Matches Task.", "success")
            form.reset()

    return render_template("tasks.html", form=form)
