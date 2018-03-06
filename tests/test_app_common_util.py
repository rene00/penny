import unittest
from app.common import util
import datetime


class TestAppCommonUtil(unittest.TestCase):

    def test_merge_dicts(self):
        self.assertEqual(
            util.merge_dicts({'a': 'b'}, {'c': 'd'}),
            {'a': 'b', 'c': 'd'}
        )

    def test_generate_transaction_hash(self):
        tx = {
            'date': datetime.datetime.strptime('2018', '%Y'),
            'debit': -1,
            'credit': 1,
            'memo': 'foo',
            'fitid': 1,
            'bankaccount_id': 1
        }
        self.assertEqual(
            util.generate_transaction_hash(**tx),
            '08834ff00ce7e87b8b8aca88e28bac7a426d989948e68b98b2f97d4e672fa4aa'
        )

        tx2 = tx.copy()
        tx2['fitid'] = None
        self.assertEqual(
            util.generate_transaction_hash(**tx2),
            '537c01bc3223a1728a38849ebb9269900a13918b7ba052c0d9ba3063450af1b3'
        )
