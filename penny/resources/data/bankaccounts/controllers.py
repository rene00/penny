from penny import models
from flask import (Blueprint, g, jsonify)
from flask_security import login_required

data_bankaccounts = Blueprint('data_bankaccounts', __name__,
                              url_prefix='/data/bankaccounts')


@data_bankaccounts.route('/')
@login_required
def bankaccounts():
    """Return data on all bankaccounts."""
    data = {'rows': []}
    for bankaccount in (models.db.session.query(models.BankAccount).
                        filter(models.BankAccount.user == g.user).all()):
        data['rows'].append(bankaccount.dump())
    data['total'] = len(data['rows'])
    return jsonify(data)
