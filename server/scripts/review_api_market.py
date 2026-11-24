"""On-demand A-share close from the official BaoStock API, with reusable 250-bar windows."""
import gzip
import json
from bisect import bisect_right
from decimal import Decimal

from build_market_review import update_window
from review_quote_sources import PRICE_FIELDS
from review_studies import calculate_evidence

FIELDS = "date,code,open,high,low,close,volume,amount,tradestatus,pctChg,preclose"


def rows(reply):
    if reply.error_code != "0":
        raise ValueError(f"BaoStock 请求失败（{reply.error_code}）")
    while reply.next():
        yield dict(zip(reply.fields, reply.get_row_data()))
    if reply.error_code != "0":
        raise ValueError(f"BaoStock 分页未完成（{reply.error_code}）")


def symbol(code):
    exchange, number = code.split(".")
    return number + "." + exchange.upper()


def quoted(row):
    return row["tradestatus"] == "1" and Decimal(row["volume"] or "0") > 0


def bar(row, factor):
    opening, high, low, close = (Decimal(row[k]) for k in PRICE_FIELDS)
    amount = Decimal(row["amount"])
    if not (0 < low <= min(opening, close) <= max(opening, close) <= high and
            amount.is_finite() and amount >= 0 and factor.is_finite() and factor > 0):
        raise ValueError(f"{row['date']} {row['code']} 日线字段无效")
    # Store only fields used for breadth, highs, patterns and sector turnover.
    return [row["date"], *(format(v.normalize(), "f") for v in (opening, high, low, close, amount, factor))]


def cold_window(bs, code, day, ipo, calendar, index, *, through=False):
    end_index = index[day]
    start = max(0, end_index - 265)
    data = []
    stop = day
    while True:
        part = list(rows(bs.query_history_k_data_plus(code, FIELDS, calendar[start], stop, "d", "3")))
        if any(r["code"] != code or not calendar[start] <= r["date"] <= stop for r in part):
            raise ValueError(f"{code} 历史行情范围错误")
        data = [r for r in part if quoted(r)] + data
        if len(data) >= 250 or calendar[start] <= ipo or start == 0:
            break
        stop = calendar[start-1]
        start = max(0, start-250)
    data = data[-250:]
    dates = [r["date"] for r in data]
    if through and not dates:
        return []
    if dates != sorted(set(dates)) or not dates or (not through and dates[-1] != day):
        raise ValueError(f"{day} {code} 缺少有效日线")
    actions = {}
    for r in rows(bs.query_adjust_factor(code, ipo, day)):
        stamp = r["dividOperateDate"]
        if r["code"] != code or stamp in actions or stamp > day:
            raise ValueError(f"{code} 复权因子日期无效")
        actions[stamp] = Decimal(r["backAdjustFactor"])
    effective = sorted(actions)
    window = []
    for r in data:
        i = bisect_right(effective, r["date"])
        window.append(bar(r, actions[effective[i-1]] if i else Decimal(1)))
    return window


def quote_window(bs, path, quote, action, previous, ipo, calendar, index, *, cached=None):
    day = quote["date"]
    old = cached
    if old is None:
        old = json.loads(gzip.decompress(path.read_bytes())) if path.exists() else []
    if old and old[-1][0] == previous:
        factor = action if action is not None else Decimal(old[-1][-1])
        if action is None and Decimal(quote["preclose"]) != Decimal(old[-1][4]):
            old = []  # Recheck corporate actions or a restated close against the historical series.
        else:
            window = (old + [bar(quote, factor)])[-250:]
    elif old and old[-1][0] == day and (action is None or action == Decimal(old[-1][-1])):
        window = old
    else:
        old = []
    if not old:
        window = cold_window(bs, quote["code"], day, ipo, calendar, index)
    if window[-1] != bar(quote, Decimal(window[-1][-1])):
        # A revised quote must also be present in the security's historical series.
        window = cold_window(bs, quote["code"], day, ipo, calendar, index)
        if window[-1] != bar(quote, Decimal(window[-1][-1])):
            raise ValueError(f"{day} {quote['code']} 横截面和历史行情不一致")
    path.write_bytes(gzip.compress(json.dumps(window, separators=(",", ":")).encode(), compresslevel=1))
    return window


def stock_evidence(row, quote, window, indices, previous):
    day, code = quote["date"], row["symbol"]
    dates = [r[0] for r in window]
    prices, adjusted, extra = [], [], {}
    for stamp, opening, high, low, close, amount, factor in window:
        h, c, f = Decimal(high), Decimal(close), Decimal(factor)
        prices.append((stamp, {code: (h, c, f)}))
        adjusted.append((stamp, h*f, c*f))
        if indices[stamp] >= indices[day]-13:
            extra[stamp] = {code: (Decimal(opening), Decimal(low), Decimal(amount)/1000)}
    update_window(row, {"close": quote["close"], "high": quote["high"], "pct_chg": quote["pctChg"]},
                  Decimal(window[-1][-1]), adjusted)
    return next(calculate_evidence(indices, {day}, {previous, day}, {code: dates}, extra,
                                   {code: row["name"]}, prices))
