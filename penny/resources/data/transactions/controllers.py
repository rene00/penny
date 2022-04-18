from penny import models
from flask import current_app as app, Blueprint, g, jsonify, request
from flask_security import auth_required
from sqlalchemy import or_
from sqlalchemy.sql import func
from datetime import datetime

data_transactions = Blueprint(
    "data_transactions", __name__, url_prefix="/data/transactions"
)


@data_transactions.route("/")
@auth_required()
def transactions():
    """Return data on all transactions."""

    data = {"rows": []}

    search = request.args.get("search")
    offset = request.args.get("offset", 1)
    limit = request.args.get("limit", 25)

    total = models.db.session.query(func.count(models.Transaction.id)).filter(
        models.Transaction.user == g.user, models.Transaction.is_deleted == 0
    )

    transactions = (
        models.db.session.query(models.Transaction)
        .filter(models.Transaction.user == g.user, models.Transaction.is_deleted == 0)
        .order_by(models.Transaction.date.desc())
    )

    if search:
        total = total.filter(models.Transaction.memo.like("%{0}%".format(search)))
        transactions = transactions.filter(
            or_(
                models.Transaction.memo.like("%{0}%".format(search)),
                models.Transaction.credit.like("{0}%".format(search)),
                models.Transaction.debit.like("-{0}%".format(search)),
            )
        )

    data["total"] = total.one()

    for transaction in transactions.offset(offset).limit(limit).all():
        data["rows"].append(transaction.dump())

    app.logger.info(data)
    return jsonify(data)


@data_transactions.route(
    "/account", defaults={"start_date": None, "end_date": None, "id": None}
)
@data_transactions.route(
    "/account/<int:id>", defaults={"start_date": None, "end_date": None}
)
@data_transactions.route("/account/<int:id>/<string:start_date>/<string:end_date>")
@auth_required()
def account(id, start_date, end_date):
    """Return data on all transactions for an account."""

    data = {"rows": []}

    search = request.args.get("search")
    offset = request.args.get("offset", 1)
    limit = request.args.get("limit", 25)

    if start_date and end_date:
        start_date = datetime.strptime(start_date, "%Y%m%d")
        end_date = datetime.strptime(end_date, "%Y%m%d")

    if start_date and end_date:
        total = models.db.session.query(func.count(models.Transaction.id)).filter(
            models.Transaction.account_id == id,
            models.Transaction.user_id == g.user.id,
            models.Transaction.date >= start_date,
            models.Transaction.date <= end_date,
            models.Transaction.is_deleted == 0,
        )
    else:
        total = models.db.session.query(func.count(models.Transaction.id)).filter(
            models.Transaction.account_id == id,
            models.Transaction.user == g.user,
            models.Transaction.is_deleted == 0,
        )

    transactions = (
        models.db.session.query(models.Transaction)
        .filter(
            models.Transaction.account_id == id,
            models.Transaction.user == g.user,
            models.Transaction.is_deleted == 0,
        )
        .order_by(models.Transaction.date.desc())
    )

    if start_date and end_date:
        transactions = transactions.filter(
            models.Transaction.date >= start_date, models.Transaction.date <= end_date
        )

    if search:
        total = total.filter(models.Transaction.memo.like("%{0}%".format(search)))
        transactions = transactions.filter(
            or_(
                models.Transaction.memo.like("%{0}%".format(search)),
                models.Transaction.credit.like("{0}%".format(search)),
                models.Transaction.debit.like("-{0}%".format(search)),
            )
        )

    data["total"] = total.one()

    for transaction in transactions.offset(offset).limit(limit).all():
        data["rows"].append(transaction.dump())

    return jsonify(data)


@data_transactions.route(
    "/bankaccount/<int:id>", defaults={"start_date": None, "end_date": None}
)
@data_transactions.route(
    "/bankaccount/<int:id>/" "<string:start_date>/<string:end_date>"
)
@auth_required()
def bankaccount(id, start_date, end_date):
    """Return data on all transactions for an bankaccount."""

    data = {"rows": []}

    search = request.args.get("search")
    offset = request.args.get("offset", 1)
    limit = request.args.get("limit", 25)

    if start_date and end_date:
        start_date = datetime.strptime(start_date, "%Y%m%d")
        end_date = datetime.strptime(end_date, "%Y%m%d")

    total = models.db.session.query(func.count(models.Transaction.id)).filter(
        models.Transaction.bankaccount_id == id,
        models.Transaction.user_id == g.user.id,
        models.Transaction.is_deleted == 0,
    )

    if start_date and end_date:
        total = total.filter(
            models.Transaction.date >= start_date, models.Transaction.date <= end_date
        )

    transactions = (
        models.db.session.query(models.Transaction)
        .filter(
            models.Transaction.bankaccount_id == id,
            models.Transaction.user == g.user,
            models.Transaction.is_deleted == 0,
        )
        .order_by(models.Transaction.date.desc())
    )

    if start_date and end_date:
        transactions = transactions.filter(
            models.Transaction.date >= start_date, models.Transaction.date <= end_date
        )

    if search:
        total = total.filter(models.Transaction.memo.like("%{0}%".format(search)))
        transactions = transactions.filter(
            or_(
                models.Transaction.memo.like("%{0}%".format(search)),
                models.Transaction.credit.like("{0}%".format(search)),
                models.Transaction.debit.like("-{0}%".format(search)),
            )
        )

    data["total"] = total.one()

    for transaction in transactions.offset(offset).limit(limit).all():
        data["rows"].append(transaction.dump())

    return jsonify(data)


@data_transactions.route(
    "/accounttype/<string:accounttype>", defaults={"start_date": None, "end_date": None}
)
@data_transactions.route(
    "/accounttype/<string:accounttype>/<string:start_date>/<string:end_date>"
)
@auth_required()
def accounttype(accounttype, start_date, end_date):
    """Return data on all transactions for an account type."""

    data = {"rows": []}

    search = request.args.get("search")
    offset = request.args.get("offset", 1)
    limit = request.args.get("limit", 25)

    if accounttype.lower() == "revenue":
        at = (
            models.db.session.query(models.AccountType)
            .filter(
                models.AccountType.name == "Revenue", models.AccountType.parent == None
            )
            .one()
        )  # noqa[E711]
    elif accounttype.lower() == "expenses":
        at = (
            models.db.session.query(models.AccountType)
            .filter(
                models.AccountType.name == "Expenses", models.AccountType.parent == None
            )
            .one()
        )  # noqa[E711]
    elif accounttype.lower() == "liabilities":
        at = (
            models.db.session.query(models.AccountType)
            .filter(
                models.AccountType.name == "Liabilities",
                models.AccountType.parent == None,
            )
            .one()
        )  # noqa[E711]
    elif accounttype.lower() == "assets":
        at = (
            models.db.session.query(models.AccountType)
            .filter(
                models.AccountType.name == "Assets", models.AccountType.parent == None
            )
            .one()
        )  # noqa[E711]
    else:
        raise Exception("Unknown accounttype")

    if start_date and end_date:
        start_date = datetime.strptime(start_date, "%Y%m%d")
        end_date = datetime.strptime(end_date, "%Y%m%d")

    results = (
        models.db.session.query(models.Transaction, models.Account, models.AccountType)
        .filter(
            models.Transaction.user_id == g.user.id,
            models.Transaction.is_deleted == 0,
        )
        .filter(models.Transaction.account_id == models.Account.id)
        .filter(models.Account.accounttype_id == models.AccountType.id)
        .filter(models.AccountType.parent == at)
        .order_by(models.Transaction.date.desc())
    )

    if start_date:
        results = results.filter(models.Transaction.date >= start_date)

    if end_date:
        results = results.filter(models.Transaction.date <= end_date)

    if search:
        results = results.filter(
            or_(
                models.Transaction.memo.like("%{0}%".format(search)),
                models.Transaction.credit.like("{0}%".format(search)),
                models.Transaction.debit.like("-{0}%".format(search)),
            )
        )

    data["total"] = results.with_entities(func.count()).scalar()

    for results in results.offset(offset).limit(limit).all():
        transaction = results[0]
        data["rows"].append(transaction.dump())

    return jsonify(data)
