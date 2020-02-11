from app import models
from flask import (Blueprint, g, jsonify, url_for)
from flask_security import login_required

data_reports = Blueprint('data_reports', __name__, url_prefix='/data/reports')


@data_reports.route('/')
@login_required
def reports():
    """Return a list of reports"""

    data = {'rows': []}

    # Add account monthly breakdown report
    data['rows'].append(
        {'report_url': '<a href="{0}">Account Monthly Breakdown<a>'.
         format(url_for('reports.account_monthly_breakdown'))}
    )

    # Add entity reports
    for entity in models.db.session.query(models.Entity) \
            .filter_by(user_id=g.user.id).order_by(models.Entity.name):
        report_url = url_for('reports.profitloss', id=entity.id,
                             report_resource='entity')
        report_name = '{0} Profit & Loss'.format(entity.name)
        data['rows'].append(
            {'report_url': '<a href="{0}">{1}</a>'.
                format(report_url, report_name)})

    # Add bankaccount reports
    for bankaccount in models.db.session.query(models.BankAccount) \
            .filter_by(user_id=g.user.id).order_by(models.BankAccount.bank):
        report_url = url_for('reports.profitloss', id=bankaccount.id,
                             report_resource='bankaccount')
        report_name = '{0.bank} {0.number} Profit & Loss'.format(bankaccount)
        data['rows'].append(
            {'report_url': '<a href="{0}">{1}</a>'.
                format(report_url, report_name)})

    data['total'] = len(data['rows'])

    return jsonify(data)
