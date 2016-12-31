from app import models
from app.models import AccountMatch, Transaction
from flask import current_app as app
from sqlalchemy.exc import IntegrityError
import re

session = models.db.session


class AccountMatchRun():
    def __init__(self, user):
        self.user = user

    @staticmethod
    def _match_transaction(accountmatch, transaction):
        transaction.account = accountmatch.account
        session.add(transaction)
        try:
            session.commit()
        except IntegrityError as e:
            app.logger.error("Failed commit: id={0.id}, error={e}".
                             format(transaction, e))
            return session.rollback()
        else:
            app.logger.info('Match transaction; transaction={0.id}, '
                            'accountmatch={1.id}'.format(
                                transaction, accountmatch))
            return transaction

    def run(self):
        # Iterate through all account matches for this user. am is
        # account match.
        for am in session.query(AccountMatch).filter_by(user=self.user):

            # Iterate through all account match regexes. amr is account
            # match regex.
            for amr in am.accountmatchfilterregexes:
                try:
                    regex = re.compile(amr.regex)
                except:
                    app.logger.error('Failed to parse regex; '
                                     'id={amr.id}'.format(amr))
                    continue
                else:
                    for transaction in session.query(Transaction). \
                            filter_by(user=self.user, account=None,
                                      bankaccount=am.bankaccount):

                        if regex.search(transaction.memo):
                            self._match_transaction(am, transaction)
