import csv
from io import StringIO
from ofxparse import OfxParser
from ofxparse.ofxparse import OfxParserException


class FileTypesError(Exception):
    pass


def is_csv(csv_file):
    with open(csv_file, 'r') as fh:
        data = csv.DictReader(fh)
    for row in data:
        field_names = ['date', 'memo', 'debit', 'credit']
        for field in row:
            if field.strip() not in field_names:
                raise FileTypesError('unknown csv field; field={}'.format(
                                     field.strip()))
            else:
                field_names.remove(field.strip())
    if len(field_names) >= 1:
        missing = ''.join('field=%s' % ''.join(f) for f in field_names)
        raise FileTypesError('missing csv field; {}'.format(missing))
    else:
        return True


def is_paypal_csv(csv_file):
    with open(csv_file, 'r') as fh:
        transactions = fh.read()
    data = csv.DictReader(StringIO(transactions))
    # these 10 fielders are found in a paypal csv file
    field_names = ['Date', 'Time', 'Time Zone', 'Name', 'Type', 'Status',
                   'Currency', 'Gross', 'Fee']

    try:
        header = data.next()
    except:
        raise FileTypesError('failed reading file')

    for key, value in header.iteritems():
        if key.strip() in field_names:
            field_names.remove(key.strip())

    if not field_names:
        return True
    else:
        return False


def is_ofx(ofx_file):
    ofx = None

    with open(ofx_file, 'r') as fh:
        try:
            ofx = OfxParser.parse(fh, fail_fast=False)
        except (OfxParserException, TypeError):
            raise FileTypesError('failed to import ofx file')

    if ofx is None:
        raise FileTypesError('unable to read ofx file')
    elif not hasattr(ofx, 'account'):
        raise FileTypesError('unable to read ofx file')
    else:
        return True


def get_bankaccount_number_from_ofx(ofx_file):
    "Return the bankaccount number from an OFX file."
    with open(ofx_file) as fh:
        ofx = OfxParser.parse(fh, fail_fast=False)
    number = '{}{}'.format(ofx.account.routing_number, ofx.account.number)
    return number
