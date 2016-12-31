"""
Proof of concept code which scrapes transactions from a credit card site
and imports them into penny.

Notes:
    - geckodriver binaries can be found at
      https://github.com/mozilla/geckodriver/releases.
    - to populate ~/.mozilla, first run firefox manually and login into
      site so that cookies file exists.
"""

import sys
import os
sys.path.insert(1, os.path.join(sys.path[0], '..'))

from app import app
from app.models import session, Transaction, BankAccount, User
from bs4 import BeautifulSoup
from flask.ext.script import Manager
from selenium import webdriver
from sqlalchemy.orm.exc import NoResultFound
from sqlalchemy.exc import IntegrityError
from xvfbwrapper import Xvfb
import time

manager = Manager(app)

import datetime
from app.common.currency import to_cents, get_credit_debit
import re


def get_credit_card_transaction(tr):
    """Takes a TR and returns a Transaction object"""
    date = tr.find(attrs={'name': 'Transaction_TransactionDate'}).string
    cardname = tr.find(attrs={'name': 'Transaction_CardName'}).string
    memo = tr.find(
        attrs={'name': 'Transaction_TransactionDescription'}).string
    amount = tr.find(attrs={'name': 'Transaction_Amount'}).string
    return CreditCardTransaction(date=date, cardname=cardname, memo=memo,
                                 amount=amount)


class CreditCardTransaction():
    def __init__(self, **kwargs):
        self.date = self._get_date(kwargs.get('date', None))
        self.cardname = self._get_cardname(kwargs.get('cardname', None))
        self.memo = self._get_memo(kwargs.get('memo', None))
        self.amount = self._get_amount(kwargs.get('amount', None))
        self.account = kwargs.get('account', None)
        self.bankaccount = kwargs.get('bankaccount', None)
        (self.credit, self.debit) = get_credit_debit(to_cents(self.amount))
        self.user = kwargs.get('user', None)
        self.fitid = None
        self.paypalid = None
        self.parent = None

    def __str__(self):
        return """
Date: {0.date}
Cardname: {0.cardname}
Memo: {0.memo}
        """.format(self)

    def __repr__(self):
        return '<CreditCardTransaction memo={0}>'.format(**self)

    @staticmethod
    def _get_date(date):
        """Return date of transaction as a datetime obj.

        Args:
            str: date of transaction.

        Notes:
            Credit Card Transaction sometimes has 'Yesterday' as the date
            string so simple logic to pick this up and convert it to the
            format I expect is required.

        Returns:
            datetime: date of transaction.
        """
        if date.lower() == 'yesterday':
            yesterday = datetime.date.today() - datetime.timedelta(1)
            date = yesterday.strftime('%d %b %Y')
        return datetime.datetime.strptime(date, '%d %b %Y')

    @staticmethod
    def _get_cardname(cardname):
        return cardname

    @staticmethod
    def _get_memo(memo):
        return re.sub(' +', ' ', memo)

    @staticmethod
    def _get_amount(amount):
        return re.sub('\$|,', '', amount)


def get_firefox_profile(profile_directory):
    """Return webdriver firefox profile obj."""
    args = {'profile_directory': profile_directory}
    return webdriver.FirefoxProfile(**args)


@manager.option('--username', dest='username')
@manager.option('--password', dest='password')
@manager.option('--bankaccount-id', dest='bankaccount_id', default=15)
@manager.option('--max-page-count', dest='max_page_count', default=5,
                required=False)
@manager.option('--user-id', dest='user_id', default=1, required=False)
@manager.option('--xvfb', dest='xvfb', default=False, required=False)
@manager.option('--debug', dest='debug', default=False, required=False)
@manager.option('--verbose', dest='verbose', default=False, required=False)
@manager.option('--geckodriver', dest='geckodriver', required=True)
@manager.option('--firefox-profile-directory',
                dest='firefox_profile_directory', required=True)
@manager.option('--login-url',
                dest='login_url', required=True)
def run(username, password, bankaccount_id, max_page_count,
        user_id, xvfb, geckodriver, firefox_profile_directory, login_url,
        **kwargs):

    debug = kwargs.get('debug')
    verbose = kwargs.get('verbose')

    if debug:
        verbose = True

    if xvfb:
        _xvfb = Xvfb(width=1280, height=720)
        _xvfb.start()

    profile = get_firefox_profile(profile_directory=firefox_profile_directory)

    driver = webdriver.Firefox(
        firefox_profile=profile, executable_path=geckodriver
    )

    driver.get(login_url)
    driver.find_element_by_id('AccessToken_Username').send_keys(username)
    driver.find_element_by_id('AccessToken_Password').send_keys(password)
    driver.find_element_by_name('SUBMIT').click()

    time.sleep(5)

    try:
        bankaccount = session.query(BankAccount). \
            filter_by(id=bankaccount_id).one()
    except NoResultFound:
        raise

    try:
        user = session.query(User).filter_by(id=user_id).one()
    except NoResultFound:
        raise

    count = 0

    while True:
        if count >= max_page_count:
            break

        driver.refresh()
        time.sleep(20)
        driver.refresh()

        soup = BeautifulSoup(driver.page_source, 'html.parser')
        table = soup.find(
            'table', attrs={'name': 'searchAllTransactionsResponse_Table'})

        for tr in table.find_all('tr', attrs={'name': 'DataContainer'}):
            credit_card_transaction = get_credit_card_transaction(tr)

            transaction = Transaction(
                bankaccount=bankaccount, user=user,
                memo=credit_card_transaction.memo,
                date=credit_card_transaction.date,
                debit=credit_card_transaction.debit,
                credit=credit_card_transaction.credit
            )
            transaction.transaction_hash = transaction.get_hash(transaction)
            session.add(transaction)
            try:
                session.commit()
            except IntegrityError:
                session.rollback()
            else:
                if verbose:
                    print(str(transaction))

        time.sleep(1)

        try:
            next = driver.find_element_by_name('nextButton')
        except:
            count = max_page_count
        else:
            count += 1
            next.click()
            time.sleep(5)

    driver.quit()
    if xvfb:
        _xvfb.stop()


if __name__ == '__main__':
    manager.run()
