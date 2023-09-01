from flask.cli import AppGroup
from penny import tasks, models
import click

task_cli: AppGroup = AppGroup("task")

@task_cli.command("fetch-tx-meta")
@click.argument("tx_id")
def task_fetch_tx_meta(tx_id) -> None:
    return tasks.fetch_tx_meta(tx_id)

@task_cli.command("tag_match")
def task_tag_match() -> None:
    user: models.User = models.User.query.filter_by(id=1).one()
    return tasks.tag_match(user.id)
