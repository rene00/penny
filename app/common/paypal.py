# Paypal _support_ was loosely based on their CSV files and it was
# fragile at best. This code hasn't been used for a while though I'll
# keep it around just incase I ever want to revisit PayPal CSV.

import csv
from StringIO import StringIO
from datetime import datetime as dt
from app.common.transactions.create import CreateTransaction

DATE_FMT = '%Y-%m-%d'


class PayPalTransaction():
    def __init__(self, date=None, type=None, tx_id=None, invoice_number=None,
                 fee=0, credit=0, debit=0, email=None, name=None,
                 bankaccount={}, memo=None):
        self.date = date
        self.type = type
        self.tx_id = tx_id
        self.invoice_number = invoice_number
        self.fee = fee
        self.debit = debit
        self.credit = credit
        self.name = name
        self.email = email
        self.bankaccount = bankaccount
        self.memo = memo

    def print_tx(self):
        print(self.date, self.type, self.tx_id, self.fee, self.debit,
              self.credit, self.email, self.name, self.memo)

    def return_as_dict(self, fee=False):
        "Return transaction dict ready to be added to API."

        txdata = {'date': self.date, 'paypalid': self.tx_id,
                  'debit': self.debit, 'credit': self.credit,
                  'bankaccount': self.bankaccount.id}

        if not fee:
            txdata['memo'] = self.memo
            return txdata
        else:
            txdata['credit'] = 0
            txdata['debit'] = self.fee
            txdata['memo'] = 'PayPal Fee - {}'.format(self.name)
            # for paypal fees paypalid must be removed or else it
            # will conflict with the main transaction.
            txdata['paypalid'] = ''
            return txdata


def get_transactions(transactions, bankaccount, user):
    reader = csv.DictReader(StringIO(transactions))
    txs = {}
    for row in reader:
        paypal = PayPalTransaction()
        paypal.bankaccount = bankaccount
        date = dt.strptime(row['Date'], '%d/%m/%Y')

        paypal.date = "{} {}".format(date.strftime("%Y-%m-%d"), row[' Time'])
        paypal.type = row[' Type']
        paypal.tx_id = row[' Transaction ID']

        # received funds
        if paypal.type == 'Payment received' or \
                paypal.type == 'Web Accept Payment Received':
            paypal.credit = abs(int(round(float(
                row[' Gross'].replace(',', ''))*100)))
            paypal.fee = int(round(float(
                row[' Fee'].replace(',', ''))*100))
            paypal.email = row[' From Email Address']
            paypal.name = row[' Name']
            paypal.memo = '{} - {}'.format(paypal.type, paypal.name)

            # if currency is not AUD ignore this. A currency conversion
            # transaction will occur next which we will pick up.
            if row[' Currency'] == 'AUD':
                data = paypal.return_as_dict()
                tx = CreateTransaction(
                    date=data['date'], debit=data['debit'],
                    credit=data['credit'], memo=data['memo'],
                    bankaccount=bankaccount, user=user,
                    paypalid=data['paypalid']).create()
                txs[tx.id] = tx

                data = paypal.return_as_dict(fee=True)
                tx = CreateTransaction(
                    date=data['date'], debit=data['debit'],
                    credit=data['credit'], memo=data['memo'],
                    bankaccount=bankaccount, user=user,
                    paypalid=data['paypalid']).create()
                txs[tx.id] = tx

        elif paypal.type == 'Refund':
            # refund
            paypal.debit = int(round(float(
                row[' Gross'].replace(',', ' '))*100))
            # refund fee is a debit and must be negative
            paypal.fee = (int(round(float(
                row[' Fee'].replace(',', ''))*100)) * -1)
            paypal.email = row[' From Email Address']
            paypal.name = row[' Name']
            paypal.memo = '{} - {}'.format(paypal.type, paypal.name)

            # if currency is not AUD ignore this. A currency conversion
            # transaction will occur next which we will pick up.
            if row[' Currency'] == 'AUD':
                # XXX: import
                data = paypal.return_as_dict()
                tx = CreateTransaction(
                    date=data['date'], debit=data['debit'],
                    credit=data['credit'], memo=data['memo'],
                    bankaccount=bankaccount, user=user,
                    paypalid=data['paypalid']).create()
                txs[tx.id] = tx

        elif paypal.type == 'Web Accept Payment Sent':
            paypal.debit = int(round(float(
                row[' Gross'].replace(',', ''))*100))
            paypal.fee = int(round(float(
                row[' Fee'].replace(',', ''))*100))
            paypal.email = row[' To Email Address']
            paypal.memo = '{} - {} - {}'.format(
                paypal.type, row[' Item Title'], paypal.email)

            # import payment sent as a transaction
            # XXX: import
            tx = CreateTransaction(
                date=data['date'], debit=data['debit'],
                credit=data['credit'], memo=data['memo'],
                bankaccount=bankaccount, user=user,
                paypalid=data['paypalid']).create()
            txs[tx.id] = tx

        elif paypal.type == 'Withdraw Funds to Bank Account':
            # intercompany transfer
            paypal.debit = int(round(float(
                row[' Gross'].replace(',', ''))*100))
            paypal.fee = abs(int(round(float(
                row[' Fee'].replace(',', ''))*100)))
            paypal.memo = paypal.type

            # import transfer sent as a transaction
            # XXX: import
            data = paypal.return_as_dict()
            tx = CreateTransaction(
                date=data['date'], debit=data['debit'],
                credit=data['credit'], memo=data['memo'],
                bankaccount=bankaccount, user=user,
                paypalid=data['paypalid']).create()
            txs[tx.id] = tx

        elif paypal.type == 'Cancelled Fee':
            # cancelled fee
            paypal.credit = int(round(float(
                row[' Gross'].replace(',', ''))*100))
            paypal.memo = paypal.type

            # import cancelled fee as a transaction
            # XXX: import
            data = paypal.return_as_dict()
            tx = CreateTransaction(
                date=data['date'], debit=data['debit'],
                credit=data['credit'], memo=data['memo'],
                bankaccount=bankaccount, user=user,
                paypalid=data['paypalid']).create()
            txs[tx.id] = tx

        elif paypal.type == 'Currency Conversion':
            # looks like received funds
            paypal.credit = abs(int(round(float(
                row[' Net'].replace(',', ''))*100)))
            if row[' Currency'] == 'AUD':
                paypal.memo = '{} - {}'.format(
                    "Payment Received", paypal.type)
                # import refund as a transaction
                data = paypal.return_as_dict()
                tx = CreateTransaction(
                    date=data['date'], debit=data['debit'],
                    credit=data['credit'], memo=data['memo'],
                    bankaccount=bankaccount, user=user,
                    paypalid=data['paypalid']).create()
                txs[tx.id] = tx
