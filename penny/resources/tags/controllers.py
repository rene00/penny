from penny import models, util
from penny.resources.tags.forms import FormTag
from flask import Blueprint, render_template, url_for, g, redirect, flash
from flask_security.decorators import auth_required
from sqlalchemy.orm.exc import NoResultFound
from sqlalchemy.exc import IntegrityError
from sqlalchemy.sql import func


tags = Blueprint("tags", __name__)


@tags.route("/tags")
@auth_required()
def _tags():
    return render_template("tags.html", data_url=url_for("data_tags.tags"))


@tags.route(
    "/tags/<int:id>",
    methods=["GET"],
)
@auth_required()
def tag(id):
    try:
        tag = (
            models.db.session.query(models.Tag)
            .filter_by(id=id, user_id=g.user.id)
            .one()
        )
    except NoResultFound:
        return url_for("accounts._accounts")

    form = FormTag(obj=tag)

    if form.validate_on_submit():
        tag.name = form.name.data
        tag.desc = form.desc.data

        models.db.session.add(tag)
        models.db.session.commit()

    transactions_amount = 0
    for transaction in tag.transactions:
        amount = 0
        if transaction.credit:
            amount = transaction.credit
        if transaction.debit:
            amount += transaction.debit
        transactions_amount += amount

    return render_template(
        "tag.html",
        form=form,
        tag=tag,
        transactions_amount=util.convert_to_float(int(transactions_amount)),
    )


@tags.route("/tags/add", methods=["GET", "POST"])
@auth_required()
def add():
    form = FormTag()
    tag = None

    if form.validate_on_submit():
        tag = models.Tag(user_id=g.user.id, name=form.name.data, desc=form.desc.data)
        models.db.session.add(tag)
        try:
            models.db.session.commit()
        except IntegrityError:
            models.db.session.rollback()
            # NOTE(rene): this flash is showing twice when it should only show once.
            flash("Failed adding tag.", "error")
            return redirect(url_for("tags.add"))

        return redirect(url_for("tags.tag", id=tag.id))

    return render_template("tag.html", form=form, tag=tag)
