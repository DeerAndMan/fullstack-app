"""Backfill missing, closed A-share sessions into the review database. Python 3.9+."""
import argparse
import calendar as month_calendar
import csv
import fcntl
import gzip
import hashlib
import io
import json
import os
import re
import shutil
import subprocess
import tarfile
import tempfile
import xml.etree.ElementTree as ET
from bisect import bisect_left, bisect_right
from collections import defaultdict, deque
from datetime import date, datetime, timedelta
from decimal import Decimal
from pathlib import Path
from zoneinfo import ZoneInfo

from build_market_review import SCOPE, SOURCE, binary_dates, daily_file, snapshot

SERVER = Path(__file__).resolve().parents[1]
RELEASES = "https://github.com/chenditc/investment_data/releases/download"
LATEST = "https://api.github.com/repos/chenditc/investment_data/releases/latest"
DOLT = "https://www.dolthub.com/api/v1alpha1/chenditc/investment_data/master"
PCF = "https://reportdocs.static.szse.cn/files/text/ETFDown"
ONE_DAY = timedelta(days=1)


def fetch(url, *options, output=None):
    args = ["curl", "--fail", "--silent", "--show-error", "--location", "--retry", "4",
            "--retry-delay", "12", "--max-time", "180", *options, url]
    if output is not None:
        subprocess.run([*args, "--output", str(output)], check=True)
    else:
        return subprocess.check_output(args)


def closed_through(now):
    local = now.astimezone(ZoneInfo("Asia/Shanghai"))
    return local.date() if local.hour >= 15 else local.date() - ONE_DAY


def month_start(end):
    previous = end.replace(day=1) - ONE_DAY
    return previous.replace(day=min(end.day, month_calendar.monthrange(previous.year, previous.month)[1])) + ONE_DAY


def plan(calendar, start, end, archived, release):
    if not calendar or end > calendar[-1]:
        raise ValueError(f"交易日历仅覆盖至 {calendar[-1:]}, 无法核对 {end}")
    due = [d for d in calendar if start <= d <= end and d not in archived]
    return [d for d in due if d <= release], [d for d in due if d > release]


def read_basket(xml, fund, day):
    root = ET.fromstring(xml)
    for tag, value in (("TradingDay", day.replace("-", "")), ("Publish", "Y"),
                       ("UnderlyingSecurityID", "399300"), ("SecurityID", fund)):
        if root.findtext("{*}" + tag) != value:
            raise ValueError(f"{day} {fund}: invalid {tag}")
    symbols = []
    for component in root.findall(".//{*}Component"):
        code = component.findtext("{*}UnderlyingSecurityID")
        exchange = component.findtext("{*}UnderlyingSecurityIDSource")
        if code == "159900" and exchange == "102":  # 申赎现金
            continue
        symbol = f"{code}.{ {'101': 'SH', '102': 'SZ'}.get(exchange, '')}"
        if not SCOPE.fullmatch(symbol):
            raise ValueError(f"{day} {fund}: unexpected PCF component {symbol}")
        symbols.append(symbol)
    if len(symbols) != 300 or len(set(symbols)) != 300:
        raise ValueError(f"{day} {fund}: expected 300 unique constituents")
    return set(symbols)


def constituents(day, work):
    pools = []
    for fund in ("159919", "159925"):
        url = f"{PCF}/pcf_{fund}_{day.replace('-', '')}.xml"
        file = work / f"pcf_{fund}_{day.replace('-', '')}.xml"
        if not file.exists():
            fetch(url, output=file)
        pools.append(read_basket(file.read_bytes(), fund, day))
    if pools[0] != pools[1]:
        raise ValueError(f"{day}: ETF 成份有分歧 {sorted(pools[0] ^ pools[1])}")
    return [(s, day, day) for s in sorted(pools[0])]


def release_info():
    tag = json.loads(fetch(LATEST))["tag_name"]
    date.fromisoformat(tag)
    manifest = json.loads(fetch(f"{RELEASES}/{tag}/qlib_bin.manifest.json"))
    if (manifest["release_tag"] != tag or date.fromisoformat(manifest["target_trade_date"]) > date.fromisoformat(tag)
            or not re.fullmatch(r"[0-9a-v]{32}", manifest["dolt_commit"])):
        raise ValueError("unexpected release manifest")
    return manifest


def extract_calendar_and_volumes(work, manifest):
    release = manifest["release_tag"]
    archive = work / "source.tar.gz"
    if not archive.exists():
        fetch(f"{RELEASES}/{release}/qlib_bin.tar.gz", output=archive)
    digest = hashlib.sha256()
    with archive.open("rb") as file:
        for block in iter(lambda: file.read(1024 * 1024), b""):
            digest.update(block)
    if ("sha256:" + digest.hexdigest() != manifest["archive_sha256"]
            or archive.stat().st_size != manifest["archive_size_bytes"]):
        raise ValueError("release archive checksum mismatch")
    with tarfile.open(archive, "r|gz") as tar:
        for member in tar:
            if member.isfile() and re.fullmatch(
                r"qlib_bin/(?:calendars/day(?:_future)?\.txt|features/(?:sz|sh)\d{6}/volume\.day\.bin)", member.name
            ):
                target = work / member.name
                target.parent.mkdir(parents=True, exist_ok=True)
                with target.open("wb") as output:
                    shutil.copyfileobj(tar.extractfile(member), output)
    return work / "qlib_bin"


def volume_history(qlib, calendar, targets):
    traded = {}
    for path in (qlib / "features").glob("*/volume.day.bin"):
        code = path.parent.name.upper()
        symbol = code[2:] + "." + code[:2]
        if not SCOPE.fullmatch(symbol):
            continue
        dates = binary_dates(path, calendar)
        first = bisect_left(dates, targets[0])
        if first < len(dates) and dates[first] <= targets[-1]:
            traded[symbol] = dates[max(0, first - 249):bisect_right(dates, targets[-1])]
    return traded


def price_partitions(work, days, commit):
    # DoltHub caps result rows and execution time. Smaller ranges recover slow queries.
    pending = deque(days[i:i + 10] for i in range(0, len(days), 10))
    while pending:
        chunk = pending.popleft()
        sql = ("select tradedate as d,left(symbol,3) as grp,json_arrayagg(json_array(symbol,high,close,adjclose)) as stocks "
               f"from ts_a_stock_eod_price as of '{commit}' where tradedate between '{chunk[0]}' and '{chunk[-1]}' "
               "group by tradedate,left(symbol,3)")
        cache = work / f"bars-{chunk[0]}.gz"
        reply = json.loads(gzip.decompress(cache.read_bytes())) if cache.is_file() else {}
        if reply.get("sql_query") != sql:
            reply = json.loads(fetch(DOLT, "-G", "--data-urlencode", "q=" + sql))
        if reply.get("query_execution_status") == "Success" and not cache.exists():
            cache.write_bytes(gzip.compress(json.dumps(reply).encode(), compresslevel=2))
        if "deadline" in reply.get("query_execution_message", "").lower() and len(chunk) > 1:
            pivot = len(chunk) // 2
            pending.extendleft((chunk[pivot:], chunk[:pivot]))
            continue
        if reply.get("sql_query") != sql or reply.get("query_execution_status") != "Success" or len(reply["rows"]) >= 1000:
            raise ValueError(f"Dolt partition failed: {chunk[0]} {reply.get('query_execution_message')}")
        bars = defaultdict(dict)
        for group in reply["rows"]:
            for symbol, high, close, adjclose in json.loads(group["stocks"], parse_float=Decimal, parse_int=Decimal):
                if not SCOPE.fullmatch(symbol):
                    continue
                rows = bars[group["d"]]
                if symbol in rows or not high >= close > 0 or not adjclose > 0:
                    raise ValueError(f"invalid bar {group['d']} {symbol}")
                # Tushare stores the source factor to four decimals and adjclose = close * factor.
                factor = (adjclose / close).quantize(Decimal(".0001"))
                if factor <= 0 or abs(adjclose / close - factor) > Decimal(".0000001"):
                    raise ValueError(f"ambiguous adjustment factor {group['d']} {symbol}")
                rows[symbol] = (high, close, factor)
        if set(bars) != set(chunk):
            raise ValueError(f"missing price dates {set(chunk) ^ bars.keys()}")
        for day in chunk:
            yield day, bars.pop(day)


def calculate(work, qlib, calendar, targets, manifest):
    source = work / "input"
    source.mkdir(exist_ok=True)
    # Download the day's provider-published percentage, including IPO and ex-rights references.
    master = source / "stocks_list.csv"
    if not master.exists():
        fetch(SOURCE + "/stocks/stocks_list.csv", output=master)
    stocks_csv = master.read_text(encoding="utf-8-sig")
    stocks = {r["symbol"]: r for r in csv.DictReader(io.StringIO(stocks_csv)) if SCOPE.fullmatch(r["symbol"])}
    if not stocks or any(not row["list_date"] for row in stocks.values()):
        raise ValueError("invalid stock master")
    memberships = {day: constituents(day, work) for day in targets}
    traded = volume_history(qlib, calendar, targets)
    earliest = min(seq[0] for seq in traded.values())
    days = calendar[calendar.index(earliest):calendar.index(targets[-1]) + 1]
    windows = defaultdict(lambda: deque(maxlen=250))
    commit = manifest["dolt_commit"]
    for day, bars in price_partitions(work, days, commit):
        for symbol, (high, close, factor) in bars.items():
            if symbol in traded and day >= traded[symbol][0]:
                windows[symbol].append((day, high * factor, close * factor))
        if day not in memberships:
            continue
        quotes = daily_file(source, "daily_kline", day)
        source_url = f"{SOURCE}/daily_kline/date/{day.replace('-', '')}.csv"
        if quotes.keys() != bars.keys():
            raise ValueError(f"{day}: 两份当日行情名单有差异 {quotes.keys() ^ bars.keys()}")
        for symbol, quote in quotes.items():
            high, close, _ = bars[symbol]
            if (high != Decimal(quote["high"]) or close != Decimal(quote["close"])
                    or Decimal(quote["vol"]) <= 0 or not Decimal(quote["pct_chg"]).is_finite()):
                raise ValueError(f"{day} {symbol}: quote verification failed")
        factors = {s: {"adj_factor": bar[2]} for s, bar in bars.items()}
        result = snapshot(day, calendar[calendar.index(day) - 1], quotes, factors, windows,
                          stocks, memberships[day], traded, days, manifest["release_tag"])
        result["source"] = source_url
        artifact = work / (day + ".json.gz")
        with gzip.open(artifact, "wb") as file:
            file.write(json.dumps(result, ensure_ascii=False, separators=(",", ":"), allow_nan=False).encode())
        yield artifact


def read_sessions(qlib, release):
    calendar = (qlib / "calendars/day.txt").read_text().splitlines()
    future = (qlib / "calendars/day_future.txt").read_text().splitlines()
    if not calendar or calendar[-1] != release or calendar != sorted(set(calendar)):
        raise ValueError("invalid release trading calendar")
    return calendar, sorted(set(calendar) | set(future))


def sync(work, start, end, importer):
    archived = set(json.loads(subprocess.check_output([str(importer), "-dates"], cwd=SERVER)))
    if start is None:
        first = month_start(end)
        if archived:
            first = min(first, date.fromisoformat(max(archived)) + ONE_DAY)
    else:
        first = start
    manifest = release_info()
    release = manifest["target_trade_date"]
    if end.isoformat() > manifest["future_end_date"]:
        raise ValueError("待核对日期超出来源日历覆盖范围")
    # Keep only a small calendar for holiday and already-published runs.
    cache = SERVER.parent / "tmp" / "market-review-calendar.json"
    saved = json.loads(cache.read_text()) if cache.is_file() else {}
    qlib = None
    if saved.get("digest") == manifest["archive_sha256"]:
        sessions = saved["sessions"]
    else:
        qlib = extract_calendar_and_volumes(work, manifest)
        calendar, sessions = read_sessions(qlib, release)
        cache.parent.mkdir(exist_ok=True, parents=True)
        with tempfile.NamedTemporaryFile(mode="w", dir=cache.parent, delete=False) as file:
            json.dump({"digest": manifest["archive_sha256"], "sessions": sessions}, file)
        Path(file.name).replace(cache)
    ready, delayed = plan(sessions, first.isoformat(), min(end.isoformat(), sessions[-1]), archived, release)
    if ready:
        if qlib is None:
            qlib = extract_calendar_and_volumes(work, manifest)
            calendar, _ = read_sessions(qlib, release)
        print(f"{ready[0]} 至 {ready[-1]}，待补 {len(ready)} 个交易日", flush=True)
        for artifact in calculate(work, qlib, calendar, ready, manifest):
            subprocess.run([str(importer), str(artifact)], cwd=SERVER, check=True)
    if delayed:
        raise RuntimeError(f"收盘行情尚未发布 {', '.join(delayed)}；下次运行继续补齐")
    after = set(json.loads(subprocess.check_output([str(importer), "-dates"], cwd=SERVER)))
    if set(ready) - after:
        raise RuntimeError("导入后仍有归档缺口")
    print(f"{first} 至 {end} 交易日归档完整，本次新增 {len(ready)} 天", flush=True)
    from review_studies import publish
    rows = json.loads(subprocess.check_output([str(importer), "-studies"], cwd=SERVER)) or []
    needed = [r for r in rows if first.isoformat() <= r["date"] <= end.isoformat()
              and not {"style","patterns","sectors"}.issubset(r["kinds"].split(","))]
    if needed:
        if qlib is None:
            qlib = extract_calendar_and_volumes(work, manifest)
            calendar, _ = read_sessions(qlib, release)
        publish(work, importer, qlib, calendar, manifest, needed)
    # Retry the most recent closed session even when its daily archive already exists.
    latest = max((d for d in after if d <= end.isoformat()), default=None)
    advice_dates = [r["date"] for r in rows if "advice-v3" not in r["kinds"].split(",")
                    and (first.isoformat() <= r["date"] <= end.isoformat() or r["date"] == latest)]
    if advice_dates:
        subprocess.run([str(importer), "-advice", *advice_dates], cwd=SERVER, check=True)



class QuotesNotPublished(ValueError):
    """The selected day is newer than the vendor's published trading data."""


def refresh_day(work, day, importer):
    """Build all inputs before publishing a selected day and its new AI advice."""
    print("REVIEW 核对交易日与上游收盘行情", flush=True)
    manifest = release_info()
    if day > manifest["target_trade_date"]:
        raise QuotesNotPublished(f"行情源最新交易日为 {manifest['target_trade_date']}，{day} 的收盘行情尚未发布")
    qlib = extract_calendar_and_volumes(work, manifest)
    calendar, _ = read_sessions(qlib, manifest["target_trade_date"])
    if day not in calendar:
        raise ValueError(f"{day} 不在数据源已发布的交易日历中")
    print("REVIEW 重算市场指标与成份名单", flush=True)
    paths = list(calculate(work, qlib, calendar, [day], manifest))
    if len(paths) != 1:
        raise ValueError(f"{day} 缺少完整行情归档")
    raw = gzip.decompress(paths[0].read_bytes())
    daily = json.loads(raw)
    if daily["date"] != day:
        raise ValueError("行情归档日期不匹配")
    # Previous-day MA50 coverage is needed for the index divergence check.
    coverage = json.loads(subprocess.check_output([str(importer), "-studies"], cwd=SERVER)) or []
    previous = next((r for r in coverage if r["date"] == daily["previous"]), {})
    values = [m["above"] for m in daily["rows"] if m.get("above") is not None]
    row = dict(date=day, batch=hashlib.sha256(raw).hexdigest(), kinds="", above=sum(values),
               eligible=len(values), prevEligible=previous.get("eligible"), prevAbove=previous.get("above"))
    print("REVIEW 重算风格、行业和个股形态", flush=True)
    from review_studies import publish
    publish(work, None, qlib, calendar, manifest, [row])
    print("REVIEW 生成当日 AI 复盘与买入评估", flush=True)
    subprocess.run([str(importer), "-refresh", str(paths[0])], cwd=SERVER, check=True)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--start", type=date.fromisoformat)
    parser.add_argument("--end", type=date.fromisoformat)
    parser.add_argument("--refresh-date", type=date.fromisoformat)
    args = parser.parse_args()
    if args.refresh_date and (args.start or args.end):
        parser.error("--refresh-date 仅接受单个交易日")
    now = datetime.now(ZoneInfo("Asia/Shanghai"))
    if args.refresh_date and now.hour < 15:
        parser.error("手动更新在北京时间 15:00 起开放")
    cutoff = closed_through(now)
    end = args.refresh_date or args.end or cutoff
    if end > cutoff or (args.start and args.start > end):
        parser.error("请选择已收盘日期范围")
    state = SERVER.parent / "tmp"
    state.mkdir(exist_ok=True)
    with (state / "market-review.lock").open("a") as lock, tempfile.TemporaryDirectory(prefix="market-review-") as folder:
        fcntl.flock(lock, fcntl.LOCK_EX)
        work = Path(folder)
        importer = work / "import-review"
        env = os.environ.copy()
        env.setdefault("GOCACHE", str(state / "go-build"))
        if installed := os.environ.get("REVIEW_IMPORTER"):
            importer = Path(installed).resolve()
        else:
            subprocess.run(["go", "build", "-o", str(importer), "./cmd/import-review"], cwd=SERVER, env=env, check=True)
        if args.refresh_date:
            try:
                refresh_day(work, end.isoformat(), importer)
            except QuotesNotPublished as error:
                print("REVIEW " + str(error), flush=True)
                raise SystemExit(75)  # Manual refresh: source data temporarily unavailable.
            except ValueError as error:
                if str(error).startswith(end.isoformat()+" 不在数据源"):
                    print("REVIEW " + str(error), flush=True)
                raise
        else:
            sync(work, args.start, end, importer)


if __name__ == "__main__":
    main()
