from app import models
from flask import (Blueprint, g, jsonify, request)
from flask.ext.security import login_required
from sqlalchemy import or_
from sqlalchemy.sql import func

data_accountmatches = Blueprint('data_accountmatches', __name__,
                                url_prefix='/data/accountmatches')


@data_accountmatches.route('/')
@login_required
def accountmatches():
    """Return data on all accountmatches."""

    data = {'rows': []}

    search = request.args.get('search')
    offset = request.args.get('offset', 1)
    limit = request.args.get('limit', 25)

    accountmatches = models.db.session.query(
        models.AccountMatch).join(models.Account).join(
            models.Entity,
            models.Account.entity_id == models.Entity.id).filter(
                models.AccountMatch.user_id == g.user.id).order_by(
                models.Entity.name,
                models.AccountMatch.name)

    if search:
        accountmatches = accountmatches.filter(
            or_(
                models.AccountMatch.name.like('%{0}%'.format(search)),
                models.Entity.name.like('%{0}%'.format(search)),
            ))
        total = accountmatches.count()
    else:
        total = models.db.session.query(
            models.AccountMatch.id).filter_by(user_id=g.user.id).count()

    data['total'] = total

    for account in accountmatches.offset(offset).limit(limit).all():
        data['rows'].append(account.dump())

    return jsonify(data)
