import unittest
from collections import deque
from decimal import Decimal
from datetime import date, datetime, timedelta, timezone

from build_market_review import snapshot
from sync_market_review import closed_through, month_start, plan, read_basket


class ReviewWindowTest(unittest.TestCase):
    def test_closed_sessions_and_publication_lag(self):
        self.assertEqual(closed_through(datetime(2026, 11, 20, 6, 59, tzinfo=timezone.utc)), date(2026, 11, 19))
        self.assertEqual(closed_through(datetime(2026, 11, 20, 7, tzinfo=timezone.utc)), date(2026, 11, 20))
        self.assertEqual(month_start(date(2026, 11, 19)), date(2026, 10, 20))
        calendar = ["2026-11-13", "2026-11-16", "2026-11-17", "2026-11-18", "2026-11-19"]
        ready, delayed = plan(calendar, calendar[0], calendar[-1], {calendar[1], calendar[3]}, calendar[3])
        self.assertEqual(ready, ["2026-11-13", "2026-11-17"])
        self.assertEqual(delayed, ["2026-11-19"])

    def test_daily_basket_identity_and_deduplication(self):
        items = "".join(
            f"<Component><UnderlyingSecurityID>{600000+i}</UnderlyingSecurityID>"
            "<UnderlyingSecurityIDSource>101</UnderlyingSecurityIDSource></Component>" for i in range(300))
        xml = ('<PCFFile xmlns="http://ts.szse.cn/Fund"><TradingDay>20261119</TradingDay>'
               '<Publish>Y</Publish><SecurityID>159919</SecurityID><UnderlyingSecurityID>399300</UnderlyingSecurityID>'
               f'<Components>{items}</Components></PCFFile>')
        self.assertEqual(len(read_basket(xml, "159919", "2026-11-19")), 300)
        for invalid in [xml.replace("20261119", "20261118"), xml.replace("600001", "600000")]:
            with self.assertRaises(ValueError):
                read_basket(invalid, "159919", "2026-11-19")

    def test_adjustment_equal_high_and_ma_denominator(self):
        # Last day's factor doubles. Its raw 5.00 equals the prior 10.00
        # on the same adjustment basis, both for the high and for the MA.
        day = "2026-10-30"
        dates = [(date.fromisoformat(day) - timedelta(days=i)).isoformat() for i in reversed(range(250))]
        source = {"600000.SH": dates, "600001.SH": dates[-49:]}
        windows = {
            symbol: deque((d, Decimal(10), Decimal(10)) for d in history)
            for symbol, history in source.items()
        }
        quotes = {symbol: {"close": "5", "high": "5", "pct_chg": "0"} for symbol in source}
        factors = {symbol: {"adj_factor": "2"} for symbol in source}
        members = [(f"{600000+i}.SH", "2024-01-01", day) for i in range(300)]
        result = snapshot(day, "2026-10-29", quotes, factors, windows, {}, members,
                          source, dates, day)["rows"]
        full, recent, absent = result[:3]
        self.assertEqual((full["previousHigh"], full["newHigh"], full["ma50"], full["above"]),
                         (5, False, 5, False))
        self.assertNotIn("newHigh", recent)
        self.assertNotIn("above", recent)
        self.assertNotIn("change", absent)
        self.assertEqual(len(result), 300)


if __name__ == "__main__":
    unittest.main()
