from penny.common.tasks import run_accountmatchrun
from penny.tasks import tag_match
from flask import Blueprint, g, flash, render_template, current_app as app
from flask_security.decorators import auth_required
from flask_wtf import FlaskForm
from wtforms import BooleanField
from rq import Queue
from redis import Redis


tasks = Blueprint("tasks", __name__)


class FormTasks(FlaskForm):
    accountmatch = BooleanField("Process Account Matches")
    tag_match = BooleanField("Process Tag Matches")


@tasks.route("/tasks", methods=["GET", "POST"])
@auth_required()
def tasks_() -> str:
    form = FormTasks()
    if form.validate_on_submit():
        q = Queue(connection=Redis.from_url(app.config["REDIS_URL"]))
        if form.accountmatch.data:
            q.enqueue(run_accountmatchrun, g.user.id)
            flash("Submitted Proccess Account Matches Task.", "success")

        if form.tag_match.data:
            q.enqueue(tag_match, g.user.id)
            flash("Submitted Proccess Account Matches Task.", "success")

    return render_template("tasks.html", form=form)
