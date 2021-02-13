from penny import models
from penny.common import forms
from flask import Blueprint, g, render_template, url_for, redirect
from flask_security import login_required
from sqlalchemy.orm.exc import NoResultFound
from penny.resources.entities.forms import FormEntity

entities = Blueprint("entities", __name__)


@entities.route("/entities")
@login_required
def _entities():
    return render_template("entities.html", data_url=url_for("data_entities.entities"))


@entities.route("/entities/<int:id>", methods=["GET", "POST"])
@login_required
def entity(id):

    try:
        entity = (
            models.db.session.query(models.Entity).filter_by(id=id, user=g.user).one()
        )
    except NoResultFound:
        return url_for("entities._entities")

    form = FormEntity(obj=entity)
    form.entitytype.choices = forms.get_entitytype_as_choices()

    if form.validate_on_submit():
        entity.name = form.name.data
        if form.entitytype.data:
            entity.entitytype_id = form.entitytype.data

        models.db.session.add(entity)
        models.db.session.commit()

    form.set_defaults(entity)

    return render_template("entity.html", form=form, entity=entity)


@entities.route("/entities/add", methods=["GET", "POST"])
@login_required
def add():
    form = FormEntity()
    form.entitytype.choices = forms.get_entitytype_as_choices()

    if form.validate_on_submit():
        entity = models.Entity(user_id=g.user.id)

        if form.name.data:
            entity.name = form.name.data

        if form.entitytype.data:
            entity.entitytype_id = form.entitytype.data

        models.db.session.add(entity)
        models.db.session.commit()

        return redirect(url_for("entities.entity", id=entity.id))

    return render_template("entity.html", form=form)
