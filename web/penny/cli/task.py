from flask.cli import AppGroup
import penny
from penny import tasks, models
import click
from rq import Queue
from redis import Redis
from penny.common.tasks import run_accountmatchrun

task_cli: AppGroup = AppGroup("task")


@task_cli.command("fetch-tx-meta")
@click.argument("tx_id")
def task_fetch_tx_meta(tx_id) -> None:
    app = penny.create_app()
    with app.app_context():
        q = Queue(
            connection=Redis.from_url(app.config["REDIS_URL"]),
            defalut_timeout=app.config.get("RQ_DEFAULT_TIMEOUT", 180),
        )
        q.enqueue(tasks.fetch_tx_meta, tx_id)


@task_cli.command("tag_match")
def task_tag_match() -> None:
    app = penny.create_app()
    with app.app_context():
        user: models.User = models.User.query.filter_by(id=1).one()
        q = Queue(
            connection=Redis.from_url(app.config["REDIS_URL"]),
            defalut_timeout=app.config.get("RQ_DEFAULT_TIMEOUT", 180),
        )
        q.enqueue(tasks.tag_match, user.id)


@task_cli.command("account_match")
def task_account_match() -> None:
    app = penny.create_app()
    with app.app_context():
        user: models.User = models.User.query.filter_by(id=1).one()
        q = Queue(
            connection=Redis.from_url(app.config["REDIS_URL"]),
            defalut_timeout=app.config.get("RQ_DEFAULT_TIMEOUT", 180),
        )
        q.enqueue(run_accountmatchrun, user.id)
