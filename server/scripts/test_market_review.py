import unittest
from datetime import date, timedelta
from decimal import Decimal
from unittest.mock import Mock, patch

from build_market_review import update_window
from review_close import csi_members


class ReviewWindowTest(unittest.TestCase):
    def test_api_membership_requires_300_distinct_effective_codes(self):
        day = "2026-11-24"
        members = [{"code": f"sh.{600000+i}", "updateDate": day} for i in range(300)]
        with patch("review_close.rows", side_effect=lambda records: iter(records)):
            bs = Mock()
            bs.query_hs300_stocks.return_value = members
            self.assertEqual(len(csi_members(bs, day)), 300)
            for invalid in (members[:-1], [*members[:-1], members[0]],
                            [*members[:-1], {**members[-1], "updateDate": "2026-11-25"}]):
                bs.query_hs300_stocks.return_value = invalid
                with self.assertRaises(ValueError):
                    csi_members(bs, day)

    def test_adjustment_equal_high_and_ma_denominator(self):
        dates = [(date(2026, 10, 30) - timedelta(days=i)).isoformat() for i in reversed(range(250))]
        window = [(d, Decimal(10), Decimal(10)) for d in dates]
        quote = {"close": "5", "high": "5", "pct_chg": "0"}
        full, recent = {"csi": True}, {"csi": True}
        update_window(full, quote, Decimal(2), window)
        update_window(recent, quote, Decimal(2), window[-49:])
        self.assertEqual((full["previousHigh"], full["newHigh"], full["ma50"], full["above"]),
                         (5, False, 5, False))
        self.assertNotIn("newHigh", recent)
        self.assertNotIn("above", recent)
