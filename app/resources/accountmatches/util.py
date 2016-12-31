from flask import current_app as app
from app.models import session, AccountMatchFilterRegex
from sqlalchemy.exc import IntegrityError


def update_filters(accountmatch, request, new_filter=False):
    for amfr in session.query(AccountMatchFilterRegex). \
            filter_by(accountmatch=accountmatch):

        # If a new filter was added, skip it or else it will be deleted.
        if new_filter and amfr == new_filter:
            continue

        regex = request.form.get('filter_regex_{0.id}'.format(amfr))
        if regex:
            amfr.regex = regex
            session.add(amfr)
        else:
            # User has removed regex to delete the amfr.
            session.delete(amfr)

        try:
            session.commit()
        except IntegrityError as e:
            app.logger.error('Error committing transaction: {0}'.format(e))
            session.rollback()


def add_filter(accountmatch, request):
    new_filter = False
    regex = request.form.get('regex')
    if regex:
        accountmatchfilterregex = AccountMatchFilterRegex(
            regex=regex, accountmatch=accountmatch)
        session.add(accountmatchfilterregex)
        session.commit()
        try:
            session.commit()
        except IntegrityError as e:
            app.logger.error('Error committing transaction: {0}'.format(e))
            session.rollback()
        else:
            new_filter = accountmatchfilterregex
    return new_filter


def update_details(accountmatch, form):
    accountmatch.name = form.name.data
    accountmatch.desc = form.desc.data
    accountmatch.account_id = form.account.data
    try:
        session.commit()
    except IntegrityError as e:
        app.logger.error('Error committing transaction: {0}'.format(e))
        session.rollback()
    finally:
        return accountmatch
