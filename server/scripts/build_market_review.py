"""Build P0 snapshots from daily prices, verified CSI membership and a second trading calendar."""
import argparse
import array
import csv
import gzip
import hashlib
import json
import re
import subprocess
import tempfile
import sys
import time
from collections import defaultdict, deque
from decimal import Decimal
from pathlib import Path

SCOPE = re.compile(r"(?:6\d{5}\.SH|(?:00|30)\d{4}\.SZ)\Z")
SOURCE = "https://money-tree.eyunzhu.com"
VERSION = "shsz-a.high250.ma50.v1"


def cached_csv(cache, key, url):
    path = cache / key
    if not path.exists():
        for attempt in range(4):
            result = subprocess.run(["curl", "-sSL", "--write-out", "%{http_code}", "--max-time", "45", url],
                                    stdout=subprocess.PIPE, check=False)
            data, status = result.stdout[:-3], result.stdout[-3:]
            if result.returncode == 0 and status == b"200":
                break
            if attempt == 3 or status not in (b"000", b"429", b"500", b"502", b"503", b"504"):
                raise RuntimeError(f"HTTP {status.decode()} while fetching {url}")
            time.sleep(30 * (attempt + 1))
        if not data.startswith(b"symbol,trade_date,"):
            raise ValueError(f"{url}: unexpected CSV")
        with tempfile.NamedTemporaryFile(dir=cache, delete=False) as tmp:
            tmp.write(data)
        Path(tmp.name).replace(path)
        time.sleep(1)
    return path


def stock_history(cache, symbol):
    result = []
    for kind in ("daily_kline/origin", "adj_factor/stock"):
        path = cached_csv(cache, f"{kind.split('/')[0]}-{symbol}.csv", f"{SOURCE}/{kind}/{symbol}.csv")
        with path.open(encoding="utf-8-sig") as file:
            rows = {}
            for row in csv.DictReader(file):
                if row["symbol"] != symbol or row["trade_date"] in rows:
                    raise ValueError(f"invalid per-stock history {path}")
                rows[row["trade_date"]] = row
            result.append(rows)
    return result


def daily_file(cache, kind, date):
    compact = date.replace("-", "")
    path = cached_csv(cache, f"{kind}-{compact}.csv", f"{SOURCE}/{kind}/date/{compact}.csv")
    with path.open(encoding="utf-8-sig") as file:
        rows = {}
        for row in csv.DictReader(file):
            if row["trade_date"] != compact:
                raise ValueError(f"wrong trading day in {path}")
            code = row["symbol"]
            if SCOPE.fullmatch(code):
                if code in rows:
                    raise ValueError(f"duplicate {date} {code}")
                rows[code] = row
        if not rows:
            raise ValueError(f"empty file {path}")
        return rows


def binary_dates(path, calendar):
    values = array.array("f")
    values.frombytes(path.read_bytes())
    if sys.byteorder != "little":
        values.byteswap()
    start = int(values[0])
    return [calendar[start + i] for i, volume in enumerate(values[1:]) if volume > 0]


def number(value):
    return float(round(value, 8))


def snapshot(date, previous, prices, factors, windows, stocks, members, traded, days, release):
    compact = date.replace("-", "")
    universe = {s for s, r in stocks.items() if r["list_date"] and r["list_date"] <= compact
                and (not r["delist_date"] or r["delist_date"] >= compact)}
    expected = {s for s, dates in traded.items() if date in dates}
    if expected != prices.keys():
        raise ValueError(f"{date}: daily quote lists disagree: {expected ^ prices.keys()}")
    csi = {s for s, start, end in members if start <= date <= end}
    if len(csi) != 300:
        raise ValueError(f"{date}: expected 300 constituents, got {len(csi)}")
    rows = []
    for symbol in sorted(universe | prices.keys() | csi):
        quote = prices.get(symbol)
        row = {"symbol": symbol, "name": stocks.get(symbol, {}).get("name") or symbol,
               "csi": symbol in csi, "bars": 0}
        rows.append(row)
        if quote is None:
            row["reason"] = "两源均无当日成交日线"
            continue
        window = windows[symbol]
        available = [d for d in traded[symbol] if d <= date][-250:]
        if [bar[0] for bar in window] != available:
            raise ValueError(f"{symbol} {date}: historical trading-day lists disagree; check inputs or extend warmup")
        close = Decimal(quote["close"])
        factor = Decimal(factors[symbol]["adj_factor"])
        row.update(change=float(quote["pct_chg"]), close=float(close), high=float(quote["high"]),
                   factor=float(factor), bars=len(window))
        if len(window) == 250:
            prior = max(list(window)[:-1], key=lambda b: b[1])
            row.update(previousHigh=number(prior[1] / factor), peakDate=prior[0],
                       newHigh=window[-1][1] > prior[1])
        else:
            row["reason"] = "有效日线不足 250 根"
        if symbol in csi and len(window) >= 50:
            total = sum(b[2] for b in list(window)[-50:])
            row.update(ma50=number(total / (50 * factor)), above=50 * close * factor > total)
    return {"date": date, "previous": previous, "rule": VERSION, "release": release,
            "source": SOURCE, "windowStart": days[0], "rows": rows}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--qlib", type=Path, required=True, help="extracted qlib_bin directory")
    parser.add_argument("--release", required=True, help="investment_data YYYY-MM-DD release")
    parser.add_argument("--csi-members", type=Path, required=True, help="verified effective memberships: symbol,from,through")
    parser.add_argument("--stocks", type=Path, required=True, help="stocks/stocks_list.csv")
    parser.add_argument("--cache", type=Path, required=True)
    parser.add_argument("--out", type=Path, required=True)
    parser.add_argument("--warmup", type=int, default=330)
    parser.add_argument("dates", nargs="+")
    args = parser.parse_args()
    dates = sorted(set(args.dates))
    calendar = (args.qlib / "calendars/day.txt").read_text().splitlines()
    if dates[-1] > args.release or args.warmup < 250 or any(d not in calendar for d in dates):
        parser.error("check release, trading dates and warmup length")
    with args.stocks.open(encoding="utf-8-sig") as file:
        stocks = {r["symbol"]: r for r in csv.DictReader(file) if SCOPE.fullmatch(r["symbol"])}
    with args.csi_members.open(encoding="utf-8-sig") as file:
        members = [(r["symbol"], r["from"], r["through"]) for r in csv.DictReader(file)]
    for date in dates:
        active = [s for s, start, end in members if start <= date <= end]
        if len(active) != 300 or len(set(active)) != 300 or not all(SCOPE.fullmatch(s) for s in active):
            parser.error(f"require 300 distinct effective constituents for {date}")
    # A second dataset verifies every target quote and the effective dates of its price window.
    first = max(0, calendar.index(dates[0]) - args.warmup + 1)
    days = calendar[first:calendar.index(dates[-1]) + 1]
    traded = {}
    for path in (args.qlib / "features").glob("*/volume.day.bin"):
        name = path.parent.name.upper()
        symbol = name[2:] + "." + name[:2]
        if SCOPE.fullmatch(symbol):
            sequence = binary_dates(path, calendar)
            # Keep enough pre-window history to detect truncated warmups.
            before = [d for d in sequence if d < days[0]][-250:]
            traded[symbol] = before + [d for d in sequence if days[0] <= d <= dates[-1]]
    args.cache.mkdir(parents=True, exist_ok=True)
    args.out.mkdir(parents=True, exist_ok=True)
    windows = defaultdict(lambda: deque(maxlen=250))
    # Only a few long-suspended stocks need bars before the usual 250-session window.
    # Their stock histories are cheaper to read than hundreds of full-market partitions.
    older = defaultdict(set)
    standard_start = calendar[max(0, calendar.index(dates[0]) - 249)]
    for symbol, sequence in traded.items():
        for target in dates:
            if target in sequence:
                for valid in [d for d in sequence if d <= target][-250:]:
                    if valid < standard_start:
                        older[valid].add(symbol)
    stock_cache = {}
    for day in days:
        compact = day.replace("-", "")
        if day < standard_start and len(older[day]) <= 3 and not all(
            (args.cache / f"{k}-{compact}.csv").is_file() for k in ("daily_kline", "adj_factor")
        ):
            prices, factors = {}, {}
            for symbol in older[day]:
                if symbol not in stock_cache:
                    stock_cache[symbol] = stock_history(args.cache, symbol)
                quotes_by_day, factors_by_day = stock_cache[symbol]
                prices[symbol], factors[symbol] = quotes_by_day[compact], factors_by_day[compact]
        else:
            prices = daily_file(args.cache, "daily_kline", day)
            factors = daily_file(args.cache, "adj_factor", day)
        for symbol, quote in prices.items():
            factor = Decimal(factors[symbol]["adj_factor"])
            high, close = Decimal(quote["high"]), Decimal(quote["close"])
            if not (factor > 0 and high >= close > 0 and Decimal(quote["vol"]) > 0):
                raise ValueError(f"invalid quote {day} {symbol}")
            windows[symbol].append((day, high * factor, close * factor))
        if day in dates:
            report = snapshot(day, calendar[calendar.index(day)-1], prices, factors, windows,
                              stocks, members, traded, days, args.release)
            payload = json.dumps(report, ensure_ascii=False, separators=(",", ":"), allow_nan=False).encode()
            path = args.out / f"{day}.json.gz"
            path.write_bytes(gzip.compress(payload, mtime=0))
            print(day, len(report["rows"]), hashlib.sha256(payload).hexdigest(), flush=True)


if __name__ == "__main__":
    main()
