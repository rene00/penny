import argparse
from ofxparse import OfxParser
from ofxparse.ofxparse import OfxParserException
from bs4 import BeautifulSoup
import sys
from StringIO import StringIO


def parse_args():
    parser = argparse.ArgumentParser()
    parser.add_argument(
        '--ofx-file', dest='ofx_file', required=True, help='OFX file')
    return parser.parse_args()


def main():
    args = parse_args()

    with open(args.ofx_file) as fh:
        t = fh.read()

    soup = BeautifulSoup(t, 'html.parser')

    try:
        ofx = OfxParser.parse(StringIO(soup))
    except OfxParserException:
        raise
    else:
        for transaction in ofx.account.statement.transactions:
            print(transaction.__dict__)

if __name__ == '__main__':
    sys.exit(main())
