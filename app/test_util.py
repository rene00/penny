import pytest
import datetime
from app.common import util


class TestUtil:

    def test_merge_dicts(self):
        assert util.merge_dicts({'a': 'b'}, {'c': 'd'}) == {'a': 'b', 'c': 'd'}

    def test_generate_transaction_hash(self):
        tx = {
            'date': datetime.datetime.strptime('2018', '%Y'),
            'debit': -1,
            'credit': 1,
            'memo': 'foo',
            'fitid': 1,
            'bankaccount_id': 1
        }
        assert util.generate_transaction_hash(**tx) == '08834ff00ce7e87b8b8aca88e28bac7a426d989948e68b98b2f97d4e672fa4aa'  # noqa[E501]

        tx2 = tx.copy()
        tx2['fitid'] = None
        assert util.generate_transaction_hash(**tx2) == '537c01bc3223a1728a38849ebb9269900a13918b7ba052c0d9ba3063450af1b3'  # noqa[E501]
