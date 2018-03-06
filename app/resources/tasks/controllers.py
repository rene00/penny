from app.common import tasks as rqtasks
from flask import (Blueprint, g, flash, render_template, request,
                   current_app as app)
from flask_security import login_required
from flask_wtf import Form
from werkzeug.datastructures import MultiDict
from wtforms import BooleanField


tasks = Blueprint('tasks', __name__)


class FormTasks(Form):
    accountmatch = BooleanField('Process Account Matches', validators=[])

    def reset(self):
        # XXX: use reset_csrf() here. See
        # https://gist.github.com/tomekwojcik/953046
        blankdata = MultiDict([])
        self.process(blankdata)


@tasks.route('/tasks', methods=['GET', 'POST'])
@login_required
def _tasks():
    form = FormTasks()

    if request.method == 'POST':
        app.logger.info('received task list; accountmatch={accountmatch}, '
                        'user={user}'.format(
                            accountmatch=form.accountmatch.data,
                            user=g.user.id))

        if form.accountmatch.data:
            rqtasks.run_accountmatchrun.delay(g.user.id)
            flash('Submitted Proccess Account Matches Task.', 'success')
            form.reset()

    return render_template('tasks.html', form=form)
