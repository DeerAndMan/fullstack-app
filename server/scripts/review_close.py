"""Publish a verified public close before its studies and historical backfill."""
import gzip
import hashlib
import json
import socket
import subprocess
import sys
import tempfile
from decimal import Decimal
from pathlib import Path
from time import perf_counter

from build_market_review import SCOPE, VERSION, update_window
from review_api_market import bar, cold_window, quote_window, quoted, rows, stock_evidence, symbol
from review_public_quotes import close_quotes
from review_quote_sources import QuotesNotPublished, SOURCES
from review_studies import record, sector_snapshot
from review_supplements import beijing_highs, style_snapshot
from sync_market_review import SERVER

CENT = Decimal(".01")
STUDIES = {"style", "patterns", "sectors", "advice-v3"}


def login_baostock(bs, stage):
    # The SDK opens its own socket. Bound that socket even when metadata comes from cache.
    previous = socket.getdefaulttimeout()
    try:
        socket.setdefaulttimeout(30)
        reply = bs.login()
    finally:
        socket.setdefaulttimeout(previous)
    if reply.error_code != "0":
        raise ValueError(f"BaoStock {stage}登录失败（{reply.error_code}，{reply.error_msg}）")


def market_roster(bs, day):
    return {symbol(r["code"]): r for r in rows(bs.query_stock_basic()) if r["ipoDate"]
            and r["ipoDate"] <= day and (not r["outDate"] or r["outDate"] >= day)
            and SCOPE.fullmatch(symbol(r["code"]))}


def published_quotes(bs, day, roster):
    daily = {}
    for row in rows(bs.query_daily_history_k_AStock(day)):
        code = symbol(row["code"])
        if code in roster and quoted(row):
            if row["date"] != day or code in daily:
                raise ValueError(f"{day} BaoStock 日线日期或代码异常 {code}")
            daily[code] = row
    active = {symbol(r["code"]) for r in rows(bs.query_all_stock(day))
              if r["tradeStatus"] == "1" and symbol(r["code"]) in roster}
    if not active or active != daily.keys():
        raise QuotesNotPublished(f"BaoStock {day} 有效成交名单与日线未齐")
    return daily


def historical_references(daily):
    return {code: (Decimal(r["preclose"]), Decimal(r["close"])) for code, r in daily.items()}


def compare_baostock(bs, day, roster, active):
    try:
        login_baostock(bs, "收盘核对")
        try:
            reference = historical_references(published_quotes(bs, day, roster))
        finally:
            bs.logout()
    except (OSError, ValueError) as exc:
        print(f"{day} BaoStock 盘后交叉核对待完成，{exc}", file=sys.stderr, flush=True)
        return
    different = active.keys() ^ reference.keys()
    for code in active.keys() & reference.keys():
        if active[code][1] != reference[code][1].quantize(CENT):
            different.add(code)
    if different:
        print(f"{day} BaoStock 盘后交叉核对，{len(different)} 只股票的成交状态或收盘价需复核", file=sys.stderr, flush=True)
    else:
        print(f"{day} BaoStock 盘后交叉核对完成，{len(active)} 只股票收盘价与有效成交名单一致", file=sys.stderr, flush=True)


def save_gzip_json(destination, value):
    destination.parent.mkdir(exist_ok=True, parents=True)
    with tempfile.NamedTemporaryFile(dir=destination.parent, delete=False) as tmp:
        path = Path(tmp.name)
        try:
            with gzip.GzipFile(fileobj=tmp, mode="wb", compresslevel=1) as file:
                file.write(json.dumps(value, separators=(",", ":")).encode())
            path.replace(destination)
        finally:
            path.unlink(missing_ok=True)


def market_inputs(bs, day, path):
    if path.exists():
        with gzip.open(path, "rt") as file:
            calendar, roster, prices = json.load(file)
        return calendar, roster, prices

    login_baostock(bs, "静态股票资料")
    try:
        calendar = [r["calendar_date"] for r in rows(bs.query_trade_dates("1990-12-19", day))
                    if r["is_trading_day"] == "1"]
        if calendar != sorted(set(calendar)) or len(calendar) < 21:
            raise ValueError("交易日历不完整")
        if calendar[-1] != day:
            return calendar, {}, {}
        roster = market_roster(bs, day)
        if not roster:
            raise ValueError("沪深上市股票名单为空")
        sessions = calendar[-21:-1]
        prices = {r["date"]: float(r["close"]) for r in rows(bs.query_history_k_data_plus(
            "sh.000300", "date,close", sessions[0], sessions[-1], "d", "3"))}
        if prices.keys() != set(sessions) or any(p <= 0 for p in prices.values()):
            raise ValueError("BaoStock 沪深300指数历史收盘不完整")
        save_gzip_json(path, (calendar, roster, prices))
        return calendar, roster, prices
    finally:
        bs.logout()


def build_core(work, importer, day, previous, roster, active, windows, csi, index, prior, source, existing=None):
    started = perf_counter()
    members = []
    amount = Decimal(0)
    changes = {}
    if not active or windows.keys() != active.keys() or not active.keys() <= roster.keys():
        raise QuotesNotPublished(f"{day} 有效成交覆盖不全")
    for code in sorted(roster.keys() | csi):
        row = dict(symbol=code, name=roster.get(code, {}).get("code_name") or code, csi=code in csi, bars=0)
        members.append(row)
        if code not in active:
            row["reason"] = "当日无有效成交日线"
            continue
        window = windows[code]
        reference, close = active[code]
        if window[-1][0] != day or Decimal(window[-1][4]) != close:
            raise QuotesNotPublished(f"{day} {code} 收盘快照和日线未齐")
        changes[code] = (close / reference - 1) * 100
        adjusted = [(r[0], Decimal(r[2])*Decimal(r[-1]), Decimal(r[4])*Decimal(r[-1])) for r in window]
        update_window(row, {"close": close, "high": window[-1][2], "pct_chg": changes[code]},
                      Decimal(window[-1][-1]), adjusted)
        amount += Decimal(window[-1][5])
    coverage = [m["above"] for m in members if "above" in m]
    if index["close"] <= index["previous20High"]:
        index["divergence"] = False
    elif prior.get("eligible") and coverage:
        index["divergence"] = sum(coverage)*prior["eligible"] < prior["above"]*len(coverage)
    report = dict(date=day, previous=previous, rule=VERSION, source=source,
                  windowStart=min(w[0][0] for w in windows.values()), release=day,
                  marketAmount=float(amount), rows=members)
    raw = json.dumps(report, ensure_ascii=False, separators=(",", ":"), allow_nan=False).encode()
    archive = work / f"{day}.json.gz"
    archive.write_bytes(gzip.compress(raw, compresslevel=1))
    batch = dict(date=day, batch=hashlib.sha256(raw).hexdigest())
    style = dict(rows=[], spread=None, index=index, marketAmount=float(amount), deferredAdvice=True)
    kinds = set()
    if existing and existing["batch"] == batch["batch"]:
        kinds = set(existing["kinds"].split(","))
    if "style" in kinds:
        style = archived_style(importer, batch)
    else:
        record(work, None, batch, "style", style)
        subprocess.run([str(importer), "-core", str(archive)], cwd=SERVER, check=True)
    print(f"REVIEW CORE {day} 核心复盘已发布，计算及归档 {perf_counter()-started:.1f} 秒", flush=True)
    return batch, members, changes, amount, style, kinds


def archived_style(importer, batch):
    return json.loads(subprocess.check_output([str(importer), "-read-style", batch["date"], batch["batch"]], cwd=SERVER))


def finish_studies(work, importer, day, calendar, indices, windows, core, refresh_supplements=True):
    batch, members, changes, amount, style, kinds = core
    pending = []
    if refresh_supplements or style.get("spread") is None:
        print("REVIEW 正在更新平均股价和近期新高指数", flush=True)
        try:
            style.update(style_snapshot(importer, day, calendar))
        except (ValueError, KeyError, OSError, subprocess.SubprocessError) as exc:
            pending.append(f"风格指数（{exc}）")
        else:
            record(work, importer, batch, "style", style)
    if refresh_supplements or style.get("beijing") is None:
        print("REVIEW 正在核对含北交的一年新高", flush=True)
        try:
            style["beijing"] = beijing_highs(importer, day, set(changes))
        except (ValueError, KeyError, ArithmeticError, OSError, subprocess.SubprocessError) as exc:
            pending.append(f"北交所新高（{exc}）")
        else:
            record(work, importer, batch, "style", style)
    if not {"patterns", "sectors"} <= kinds:
        print("REVIEW 正在补充个股形态与行业容量", flush=True)
        finish_market_studies(work, importer, day, calendar, indices, windows, batch, members, changes, amount, kinds)
    if "advice-v3" not in kinds:
        print("REVIEW 正在生成当日 AI 复盘建议", flush=True)
        subprocess.run([str(importer), "-advice", day], cwd=SERVER, check=True)
    return pending


def finish_market_studies(work, importer, day, calendar, indices, windows, batch, members, changes, amount, kinds):
    previous = calendar[indices[day]-1]
    eligible, hits, patterns = set(), set(), []
    for m in members:
        code = m["symbol"]
        if code not in windows:
            continue
        exchange_code = code[-2:].lower() + "." + code[:6]
        last = windows[code][-1]
        quote = dict(date=day, code=exchange_code, close=last[4], high=last[2], pctChg=changes[code])
        _, eligible_s, hits_s, patterns_s, _ = stock_evidence(m.copy(), quote, windows[code], indices, previous)
        eligible.update(eligible_s)
        hits.update(hits_s)
        patterns.extend(patterns_s)
    if "patterns" not in kinds:
        record(work, importer, batch, "patterns", dict(rows=patterns))
    if "sectors" not in kinds:
        print("REVIEW 正在核对各行业的成份及二十日涨幅", flush=True)
        sectors = sector_snapshot(day, calendar, eligible, hits)
        sectors["marketAmount"] = float(amount)
        record(work, importer, batch, "sectors", sectors)


def rewind_windows(bs, cache, day, daily, roster, calendar, indices, actions):
    """Rebuild overlapping older windows from dated market API pages, preserving the rolling cache."""
    completed, pending = {}, {}
    for code, quote in daily.items():
        path = cache / f"{code}.gz"
        if not path.exists():
            continue
        with gzip.open(path, "rt") as file:
            cached = json.load(file)
        if not cached or cached[-1][0] <= day:
            continue
        window = [r for r in cached if r[0] <= day]
        if not window or window[-1] != bar(quote, actions.get(code, Decimal(window[-1][-1]))):
            continue
        if len(cached) < 250 or window[0][0] <= roster[code]["ipoDate"] or len(window) == 250:
            completed[code] = window
        elif len(window) >= 225:
            pending.setdefault(indices[window[0][0]] - 1, {})[code] = window

    while pending:
        i = max(pending)
        group = pending.pop(i)
        # Sparse dates cost fewer requests through the per-stock historical endpoint.
        if i < 0 or len(group) < 12:
            continue
        stamp, following = calendar[i:i+2]
        print(f"REVIEW {day} 正在按交易日补齐历史窗口，{stamp} 共 {len(group)} 只", flush=True)
        changed = {symbol(r["code"]) for r in rows(bs.query_daily_adjust_factor(following))
                   if r["dividOperateDate"] == following}
        matched = set()
        for r in rows(bs.query_daily_history_k_AStock(stamp)):
            code = symbol(r["code"])
            if code not in group or code in changed or not quoted(r):
                continue
            if r["date"] != stamp or code in matched:
                raise ValueError(f"{stamp} 回溯日线日期或代码异常 {code}")
            matched.add(code)
            window = group[code]
            window.insert(0, bar(r, Decimal(window[0][-1])))
            if len(window) == 250 or stamp == roster[code]["ipoDate"]:
                completed[code] = window
            else:
                pending.setdefault(i - 1, {})[code] = window
        # An action, suspension, or missing quote goes through the full security-level check.
    if completed:
        print(f"REVIEW {day} 已回溯核对 {len(completed)} 只历史窗口，其余按个股日线核验", flush=True)
    return completed


def roll_published_windows(bs, work, importer, target, roster, calendar, indices, day, daily):
    actions = {symbol(r["code"]): Decimal(r["backAdjustFactor"]) for r in rows(bs.query_daily_adjust_factor(day))
               if r["dividOperateDate"] == day}
    if target.exists():
        with gzip.open(target, "rt") as file:
            prepared = json.load(file)
        if prepared.keys() == daily.keys() and all(
            w and w[-1] == bar(daily[code], actions.get(code, Decimal(w[-1][-1])))
            for code, w in prepared.items()
        ):
            print(f"REVIEW {day} 复用 {len(prepared)} 只股票的已核验历史窗口", flush=True)
            return prepared

    cache = work / "windows"
    cache.mkdir()
    subprocess.run([str(importer), "-read-quote-cache", str(cache)], cwd=SERVER, check=True)
    prepared = rewind_windows(bs, cache, day, daily, roster, calendar, indices, actions)
    pending = []
    previous = calendar[indices[day]-1]
    for code, quote in sorted(daily.items()):
        if code in prepared:
            continue
        path = cache / f"{code}.gz"
        # A historical backfill must not regress the rolling cache for a future close.
        if path.exists():
            with gzip.open(path, "rt") as file:
                cached = json.load(file)
        else:
            cached = []
        latest = cached[-1][0] if cached else ""
        windows = quote_window(bs, path, quote, actions.get(code), previous, roster[code]["ipoDate"], calendar, indices,
                               cached=cached)
        prepared[code] = windows
        if latest <= day:
            pending.append(str(path))
        if len(pending) >= 128:
            subprocess.run([str(importer), "-write-quote-cache", *pending], cwd=SERVER, check=True)
            pending.clear()
        if len(prepared) % 128 == 0:
            print(f"REVIEW {day} 历史日线已核验 {len(prepared)}/{len(daily)} 家", flush=True)
    if pending:
        subprocess.run([str(importer), "-write-quote-cache", *pending], cwd=SERVER, check=True)
    save_gzip_json(target, prepared)
    return prepared


def prepare_windows(bs, work, importer, target, roster, calendar, indices):
    """Warm from BaoStock's published previous session and reuse the rolling DB cache."""
    if target.exists():
        with gzip.open(target, "rt") as file:
            return json.load(file)
    previous = calendar[-2]
    login_baostock(bs, "历史行情")
    try:
        daily = published_quotes(bs, previous, roster)
        return roll_published_windows(bs, work, importer, target, roster, calendar, indices, previous, daily)
    finally:
        bs.logout()


def csi_members(bs, day):
    members = list(rows(bs.query_hs300_stocks(date=day)))
    codes = {symbol(r["code"]) for r in members}
    if (len(members) != 300 or len(codes) != 300 or not all(SCOPE.fullmatch(c) for c in codes)
            or any(not r["updateDate"] or r["updateDate"] > day for r in members)):
        raise ValueError(f"{day} 沪深300成份名单无效")
    return codes


def sync_baostock(work, state, end, importer, *, core_only=False):
    import baostock as bs

    started = perf_counter()
    day = end.isoformat()
    state = state / "review-public-windows"
    calendar, roster, prices = market_inputs(bs, day, state / f"roster-{day}.json.gz")
    if day not in calendar:
        raise ValueError(f"{day} 不是交易日")
    indices = {stamp: i for i, stamp in enumerate(calendar)}
    previous = calendar[-2]
    login_baostock(bs, "历史行情")
    try:
        print(f"REVIEW {day} 核对 BaoStock 收盘日线与成交名单", flush=True)
        daily = published_quotes(bs, day, roster)
        csi = csi_members(bs, day)
        prices_today = list(rows(bs.query_history_k_data_plus("sh.000300", "date,close", day, day, "d", "3")))
        if len(prices_today) != 1 or prices_today[0]["date"] != day:
            raise QuotesNotPublished(f"BaoStock {day} 沪深300收盘尚未齐备")
        csi_close = Decimal(prices_today[0]["close"])
        if not csi_close.is_finite() or csi_close <= 0:
            raise ValueError(f"{day} 沪深300收盘价无效")
        windows = roll_published_windows(bs, work, importer, state / f"{day}.json.gz",
                                         roster, calendar, indices, day, daily)
    finally:
        bs.logout()

    coverage = json.loads(subprocess.check_output([str(importer), "-studies"], cwd=SERVER)) or []
    prior = next((r for r in coverage if r["date"] == previous), {})
    index = dict(close=float(csi_close), previous=prices[previous], previous20High=max(prices.values()), divergence=None)
    print(f"REVIEW {day} 收盘行情和历史窗口已核对，耗时 {perf_counter()-started:.1f} 秒", flush=True)
    core = build_core(work, importer, day, previous, roster, historical_references(daily), windows, csi, index,
                      prior, SOURCES["baostock"][1], next((r for r in coverage if r["date"] == day), None))
    (state / f"{day}.manual").unlink(missing_ok=True)
    if not core_only:
        complete_close(work, importer, end, calendar, indices, windows, core)


def complete_close(work, importer, end, calendar, indices, windows, core, verify=None, refresh_supplements=True):
    day = end.isoformat()
    pending = finish_studies(work, importer, day, calendar, indices, windows, core, refresh_supplements)
    if verify is not None and not core[-1]:
        print("REVIEW 正在与 BaoStock 交叉核对收盘行情", flush=True)
        verify()
    if pending:
        print("REVIEW 以下数据待重试：" + "、".join(pending), flush=True)
        raise SystemExit(77)


def append_close(day, roster, active, snapshots, prepared):
    windows = {}
    for code, (reference, _) in active.items():
        window = prepared.get(code, [])
        if not window and roster[code]["ipoDate"] != day:
            raise QuotesNotPublished(f"{code} 未找到前一笔有效成交日线")
        history = window[-249:]
        factor = Decimal(1)
        if window:
            last_close, factor = Decimal(window[-1][4]), Decimal(window[-1][-1])
            if last_close != reference:
                # Use a common exact scale across the window so an equal high stays equal.
                history = [[*bar[:-1], str(Decimal(bar[-1]) * reference)] for bar in history]
                factor *= last_close
        windows[code] = history + [[*snapshots[code], str(factor)]]
    return windows


def sync_public(work, state, end, importer, *, warm_only=False, core_only=False, scheduled=False):
    import baostock as bs

    started = perf_counter()
    day = end.isoformat()
    state = state / "review-public-windows"
    if scheduled and (state / f"{day}.manual").exists():
        return False
    calendar, roster, prices = market_inputs(bs, day, state / f"roster-{day}.json.gz")
    if day not in calendar:
        if scheduled:
            return False
        raise ValueError(f"{day} 不是交易日")
    indices = {d: i for i, d in enumerate(calendar)}
    previous = calendar[-2]
    cache = state / f"{previous}.json.gz"
    if warm_only:
        prepare_windows(bs, work, importer, cache, roster, calendar, indices)
        for pattern in ("????-??-??.json.gz", "roster-????-??-??.json.gz"):
            for old in sorted(state.glob(pattern))[:-2]:
                old.unlink(missing_ok=True)
        return

    coverage = json.loads(subprocess.check_output([str(importer), "-studies"], cwd=SERVER)) or []
    archived = {r["date"]:r for r in coverage}
    if scheduled and STUDIES <= set(archived.get(day, {}).get("kinds", "").split(",")):
        study = archived_style(importer, archived[day])
        if study.get("beijing") is not None and study.get("spread") is not None:
            return False
    print(f"REVIEW {day} 核对上一交易日的历史窗口", flush=True)
    prepared = prepare_windows(bs, work, importer, cache, roster, calendar, indices)

    print(f"REVIEW {day} 正在读取并校验沪深收盘报价", flush=True)
    active, snapshots, csi_close, source = close_quotes(day, sorted(roster))
    missing = [code for code in active if code not in prepared and roster[code]["ipoDate"] < day]
    print(f"REVIEW {day} 正在核对沪深300成份及 {len(missing)} 只股票的历史窗口", flush=True)
    login_baostock(bs, "沪深300成份与停复牌历史")
    try:
        csi = csi_members(bs, day)
        for code in missing:
            exchange_code = code[-2:].lower() + "." + code[:6]
            prepared[code] = cold_window(bs, exchange_code, previous, roster[code]["ipoDate"], calendar, indices,
                                         through=True)
    finally:
        bs.logout()
    windows = append_close(day, roster, active, snapshots, prepared)
    index = dict(close=float(csi_close), previous=prices[previous], previous20High=max(prices.values()),
                 divergence=None)
    print(f"REVIEW {day} 收盘覆盖和历史窗口已核对，耗时 {perf_counter()-started:.1f} 秒", flush=True)
    core = build_core(work, importer, day, previous, roster, active, windows, csi, index, archived.get(previous, {}), source, archived.get(day))
    (state / f"{day}.manual").unlink(missing_ok=True)
    if core_only:
        return

    complete_close(work, importer, end, calendar, indices, windows, core,
                   verify=lambda: compare_baostock(bs, day, roster, active), refresh_supplements=not scheduled)
