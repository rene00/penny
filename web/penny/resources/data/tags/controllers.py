from penny import models
from flask import Blueprint, g, jsonify, request
from flask_security.decorators import auth_required
from sqlalchemy.sql import func

data_tags = Blueprint("data_tags", __name__, url_prefix="/data/tags")


@data_tags.route("/")
@auth_required()
def tags():
    """Return data on all tags."""

    data = {"rows": []}

    search = request.args.get("search")
    offset = request.args.get("offset", 1)
    limit = request.args.get("limit", 25)

    total = models.db.session.query(func.count(models.Tag.id)).filter_by(
        user_id=g.user.id
    )

    tags = (
        models.db.session.query(models.Tag)
        .filter(models.Tag.user_id == g.user.id)
        .order_by(models.Tag.name)
    )

    if search:
        total = total.filter(models.Tag.name.like("%{0}%".format(search)))
        tags = tags.filter(models.Tag.name.like("%{0}%".format(search)))

    data["total"] = total.one()[0]

    for tag in tags.offset(offset).limit(limit).all():
        data["rows"].append(tag.dump())

    return jsonify(data)
