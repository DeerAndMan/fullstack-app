import gzip
import json
import tempfile
import unittest
from datetime import date, timedelta
from decimal import Decimal
from pathlib import Path
from unittest.mock import Mock, patch

from review_api_market import bar, quote_window, stock_evidence
from review_quote_sources import QuotesNotPublished
from review_public_quotes import close_quotes, parse_quote
from review_close import append_close, build_core, login_baostock, rewind_windows


class MarketAPITest(unittest.TestCase):
    def test_baostock_login_bounds_its_socket_and_reports_the_stage(self):
        import socket
        original = socket.getdefaulttimeout()
        bs = Mock()

        def failed_login():
            self.assertEqual(socket.getdefaulttimeout(), 30)
            return Mock(error_code="10002007", error_msg="网络接收错误")

        bs.login.side_effect = failed_login
        with self.assertRaisesRegex(ValueError, "历史行情登录失败.*10002007"):
            login_baostock(bs, "历史行情")
        self.assertEqual(socket.getdefaulttimeout(), original)

    def test_rewind_trims_future_bars_and_defers_ex_rights_for_full_history(self):
        days = [(date(2025, 1, 1) + timedelta(days=n)).isoformat() for n in range(251)]
        target = days[-2]
        codes = [f"600{i:03d}" for i in range(12)]
        history = [[d, "10", "10", "10", "10", "100", "1"] for d in days[1:]]
        daily, roster, prefix = {}, {}, []
        bs = Mock()
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            for number in codes:
                code = number + ".SH"
                quote = dict(date=target, code="sh." + number, open="10", high="10", low="10",
                             close="10", amount="100", volume="10", tradestatus="1")
                daily[code] = quote
                roster[code] = {"ipoDate": "2000-01-01"}
                prefix.append(dict(quote, date=days[0]))
                (root / f"{code}.gz").write_bytes(gzip.compress(json.dumps(history).encode()))
            changed = [dict(code="sh." + codes[0], dividOperateDate=days[1])]
            with patch("review_close.rows", side_effect=[changed, prefix]):
                result = rewind_windows(bs, root, target, daily, roster, days,
                                        {d: i for i, d in enumerate(days)}, {})
            self.assertEqual(set(result), set(daily) - {codes[0] + ".SH"})
            for window in result.values():
                self.assertEqual([bar[0] for bar in window], days[:-1])
            stored = json.loads(gzip.decompress((root / f"{codes[1]}.SH.gz").read_bytes()))
            self.assertEqual(stored, history)

    def test_public_quote_requires_close_and_preserves_amount_in_yuan(self):
        day = "2026-11-24"
        fields = [""] * 33
        fields[1:10] = ["10.62", "10.75", "10.66", "10.71", "10.45", "0", "0", "32658350", "347352876.000"]
        fields[30:33] = [day, "14:59:59", "00"]
        with self.assertRaises(ValueError):
            parse_quote("sina", fields, day)
        fields[31] = "15:00:00"
        reference, bar = parse_quote("sina", fields, day)
        self.assertEqual(reference, Decimal("10.75"))
        self.assertEqual(Decimal(bar[5]), Decimal(347352876))
        fields[1:10] = ["0", "10.75", "0", "0", "0", "0", "0", "0", "0"]
        fields[32] = "03"
        self.assertIsNone(parse_quote("sina", fields, day))

    def test_public_quote_falls_back_when_primary_is_unavailable(self):
        day = "2026-11-24"
        fields = ["0"] * 33
        fields[1:10] = ["10.50", "10", "11", "11", "10.50", "0", "0", "10000", "110000"]
        fields[30:33] = [day, "15:00:00", "00"]
        code = "600000.SH"

        def feed(provider, codes):
            if provider == "tencent":
                raise ValueError("missing")
            return {k: fields for k in codes}

        with patch("review_public_quotes.fetch_rows", side_effect=feed):
            quotes, bars, index, source = close_quotes(day, [code])
        self.assertEqual(quotes[code], (Decimal(10), Decimal(11)))
        self.assertEqual(bars[code][-1], "110000")
        self.assertEqual(index, Decimal(11))
        self.assertIn("sina.com.cn", source)

    def test_public_close_uses_exchange_reference_for_ex_rights(self):
        prior, day, code = "2026-11-20", "2026-11-23", "600000.SH"
        history = {code: [[prior, "10", "10", "10", "10", "100", "1"]]}
        raw = {code: [day, "5", "5", "5", "5", "200"]}
        windows = append_close(day, {code: {"ipoDate": prior}}, {code: (Decimal(5), Decimal(5))}, raw, history)
        first, last = windows[code]
        self.assertEqual(Decimal(first[2]) * Decimal(first[-1]), Decimal(last[2]) * Decimal(last[-1]))

    def test_public_core_refuses_missing_coverage_before_publication(self):
        day = "2026-11-23"
        tickers = {"600000.SH": (Decimal(10), Decimal(11))}
        with patch("review_close.subprocess.run") as commit:
            with self.assertRaises(QuotesNotPublished):
                build_core(None, None, day, "2026-11-20", {"600000.SH": {}}, tickers, {}, set(), {}, {}, "source")
        commit.assert_not_called()

    def test_public_publishes_core_without_waiting_for_history(self):
        import sys
        from contextlib import ExitStack
        from types import SimpleNamespace
        from review_close import sync_public

        end = date(2026, 11, 23)
        days = [(end - timedelta(days=n)).isoformat() for n in range(20, -1, -1)]
        prices = {d: 10.0 for d in days[:-1]}
        events = []

        def fail_secondary(*args):
            events.append("secondary")
            raise RuntimeError("AI unavailable")

        with tempfile.TemporaryDirectory() as tmp, ExitStack() as stack:
            work = Path(tmp)
            stack.enter_context(patch.dict(sys.modules, baostock=SimpleNamespace(login=lambda: SimpleNamespace(error_code="0"), logout=lambda: None)))
            stack.enter_context(patch("review_close.market_inputs", return_value=(days, {}, prices)))
            stack.enter_context(patch("review_close.prepare_windows", return_value={}))
            stack.enter_context(patch("review_close.subprocess.check_output", return_value=b"[]"))
            stack.enter_context(patch("review_close.close_quotes", return_value=({}, {}, Decimal(10), "source")))
            stack.enter_context(patch("review_close.csi_members", return_value=[]))
            stack.enter_context(patch("review_close.build_core", side_effect=lambda *args: events.append("core")))
            stack.enter_context(patch("review_close.finish_studies", side_effect=fail_secondary))
            with self.assertRaisesRegex(RuntimeError, "AI unavailable"):
                sync_public(work, work, end, work/"importer")
        self.assertEqual(events, ["core", "secondary"])

    def test_day_reserved_for_manual_rebuild_skips_scheduled_collection(self):
        import sys
        from review_close import sync_public
        with tempfile.TemporaryDirectory() as tmp, patch.dict(sys.modules, baostock=object()), patch("review_close.market_inputs") as collect:
            work = Path(tmp)
            (work / "review-public-windows").mkdir()
            (work / "review-public-windows/2026-11-24.manual").touch()
            self.assertFalse(sync_public(work, work, date(2026, 11, 24), None, scheduled=True))
            collect.assert_not_called()

    def test_repeat_update_reuses_complete_same_batch_studies(self):
        from review_close import finish_studies
        kinds = {"patterns", "sectors", "advice-v3", "style"}
        core = ({"date": "2026-11-24"}, [], {}, 0, {"spread": 0, "beijing": {"highs": 0}}, kinds)
        with patch("review_close.finish_market_studies") as update, patch("review_close.subprocess.run") as ai:
            self.assertFalse(finish_studies(None, None, "2026-11-24", [], {}, {}, core, refresh_supplements=False))
            update.assert_not_called()
            ai.assert_not_called()

    def test_manual_refresh_updates_both_even_for_an_existing_batch(self):
        from review_close import finish_studies
        style = {"spread": 1, "beijing": {"highs": 2}}
        core = ({"date": "2026-11-24"}, [], {"600000.SH": 1}, 0, style, {"patterns", "sectors", "advice-v3"})
        with patch("review_close.style_snapshot", return_value={"spread": 0}), \
             patch("review_close.beijing_highs", return_value={"highs": 0}) as fetch, patch("review_close.record"):
            self.assertEqual(finish_studies(None, None, "2026-11-24", [], {}, {}, core), [])
        fetch.assert_called_once_with(None, "2026-11-24", {"600000.SH"})
        self.assertEqual(style, {"spread": 0, "beijing": {"highs": 0}})

    def test_incomplete_beijing_keeps_verified_fields_and_reports_partial(self):
        from review_close import finish_studies
        style = {"spread": None, "index": {"divergence": False}}
        core = ({"date": "2026-11-24"}, [], {}, 0, style, {"patterns", "sectors", "advice-v3"})
        with patch("review_close.style_snapshot", return_value={"spread": 1.5}), \
             patch("review_close.beijing_highs", side_effect=ValueError("缺少当日行情")), \
             patch("review_close.record") as save:
            pending = finish_studies(None, None, "2026-11-24", [], {}, {}, core)
        self.assertEqual(pending, ["北交所新高（缺少当日行情）"])
        save.assert_called_once()
        self.assertEqual(style, {"spread": 1.5, "index": {"divergence": False}})

    def test_beijing_adjusted_equal_high_and_short_history(self):
        from review_supplements import beijing_window
        rows = [dict(date=(date(2026, 1, 1) + timedelta(days=i)).isoformat(), close=10, high=10)
                for i in range(250)]
        final = rows[-1]
        final.update(close=5, high=5)
        snapshot = dict(ts_code="920118.BJ", trade_date=final["date"].replace("-", ""), close=5, high=5)
        data = dict(bars=rows, actions=[dict(Time=final["date"], Peigu=0, Peigujia=0, Fenhong=0, Songzhuangu=10)])
        self.assertEqual(beijing_window(snapshot, data), (True, False))
        snapshot["high"] = final["high"] = 6
        self.assertEqual(beijing_window(snapshot, data), (True, True))
        rows.pop(0)
        self.assertEqual(beijing_window(snapshot, data), (False, False))

    def test_warm_cache_rolls_and_uses_effective_factor(self):
        # A split halves the unadjusted price but ties the previous adjusted high.
        first = date(2025, 1, 1)
        sessions = [(first + timedelta(days=i)).isoformat() for i in range(251)]
        quote = dict(date=sessions[-1], code="sh.600000", open="5.00", high="5.00", low="5", close="5",
                     preclose="5", pctChg="0", amount="1234.50")
        history = [[d,"10","10","10","10","100","1"] for d in sessions[:-1]]
        with tempfile.TemporaryDirectory() as folder, patch("review_api_market.cold_window") as cold:
            path = Path(folder) / "600000.SH.gz"
            path.write_bytes(gzip.compress(json.dumps(history).encode()))
            indices = {d:i for i,d in enumerate(sessions)}
            window = quote_window(None, path, quote, Decimal(2), sessions[-2], sessions[0], sessions, indices)
            self.assertEqual((window[0][0],window[-1]), (sessions[1],bar(quote, Decimal(2))))
            self.assertEqual(quote_window(None, path, quote, Decimal(2), sessions[-2], sessions[0], sessions, indices), window)
            member = dict(symbol="600000.SH", name="浦发银行", csi=True)
            _, eligible, hits, _, amount = stock_evidence(member, quote, window, indices, sessions[-2])
            self.assertEqual((member["bars"],member["above"],member["newHigh"]), (250,False,False))
            self.assertEqual((eligible,hits,amount), ({"600000.SH"},set(),1234.5))
            cold.assert_not_called()

    def test_published_quotes_rejects_an_incomplete_traded_roster(self):
        from types import SimpleNamespace
        from review_close import published_quotes

        day = "2026-11-23"
        active = "sh.600000"
        missing = "sh.600001"
        roster = {code[3:] + ".SH": {} for code in (active, missing)}
        bs = SimpleNamespace(
            query_daily_history_k_AStock=lambda _: [dict(date=day, code=active, tradestatus="1", volume="10")],
            query_all_stock=lambda _: [dict(code=code, tradeStatus="1") for code in (active, missing)],
        )
        with patch("review_close.rows", side_effect=iter):
            with self.assertRaises(QuotesNotPublished):
                published_quotes(bs, day, roster)

    def test_historical_core_reuses_the_prepared_day_without_exporting_rolling_cache(self):
        from types import SimpleNamespace
        from review_close import roll_published_windows

        day = "2026-11-23"
        code = "600000.SH"
        quote = dict(date=day, code="sh.600000", open="10", high="11", low="10", close="11", amount="1000")
        expected = {code: [bar(quote, Decimal(2))]}
        bs = SimpleNamespace(query_daily_adjust_factor=lambda _: [
            dict(code="sh.600000", dividOperateDate=day, backAdjustFactor="2"),
        ])
        with tempfile.TemporaryDirectory() as tmp:
            work = Path(tmp)
            target = work / "dated-window.json.gz"
            target.write_bytes(gzip.compress(json.dumps(expected).encode()))
            with patch("review_close.rows", side_effect=iter), patch("review_close.subprocess.run") as importer:
                actual = roll_published_windows(bs, work, "importer", target, {}, [], {}, day, {code: quote})
            self.assertEqual(actual, expected)
            importer.assert_not_called()

    def test_historical_backfill_does_not_replace_a_later_rolling_window(self):
        from types import SimpleNamespace
        from review_close import roll_published_windows

        days = ["2026-11-20", "2026-11-23", "2026-11-24"]
        code = "600000.SH"
        quote = dict(date=days[1], code="sh.600000", open="10", high="11", low="10", close="11", amount="1000")
        later = bar({**quote, "date": days[2]}, Decimal(1))
        target_row = bar(quote, Decimal(1))
        bs = SimpleNamespace(query_daily_adjust_factor=lambda _: [])

        def export(argv, **kwargs):
            self.assertEqual(argv[1], "-read-quote-cache")
            (Path(argv[2]) / f"{code}.gz").write_bytes(gzip.compress(json.dumps([later]).encode()))

        with tempfile.TemporaryDirectory() as tmp:
            work = Path(tmp)
            with patch("review_close.rows", side_effect=iter), \
                    patch("review_close.subprocess.run", side_effect=export) as importer, \
                    patch("review_api_market.cold_window", return_value=[target_row]):
                actual = roll_published_windows(bs, work, "importer", work/"day.gz", {code: {"ipoDate": days[0]}},
                                                days, {d: i for i, d in enumerate(days)}, days[1], {code: quote})
            self.assertEqual(actual, {code: [target_row]})
            self.assertEqual(importer.call_count, 1)
