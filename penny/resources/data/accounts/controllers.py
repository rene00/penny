from penny import models
from flask import Blueprint, g, jsonify, request
from flask_security import auth_required
from sqlalchemy.sql import func

data_accounts = Blueprint("data_accounts", __name__, url_prefix="/data/accounts")


@data_accounts.route("/")
@auth_required()
def accounts():
    """Return data on all accounts."""

    data = {"rows": []}

    search = request.args.get("search")
    offset = request.args.get("offset", 1)
    limit = request.args.get("limit", 25)

    total = models.db.session.query(func.count(models.Account.id)).filter_by(
        user_id=g.user.id
    )

    accounts = (
        models.db.session.query(models.Account)
        .join(models.Entity)
        .filter(models.Account.user_id == g.user.id)
        .order_by(models.Entity.name, models.Account.name)
    )

    if search:
        total = total.filter(models.Account.name.like("%{0}%".format(search)))
        accounts = accounts.filter(models.Account.name.like("%{0}%".format(search)))

    data["total"] = total.one()

    for account in accounts.offset(offset).limit(limit).all():
        data["rows"].append(account.dump())

    return jsonify(data)
