"""Same-session public A-share quotes, with a complete-provider fallback."""
import re
import subprocess
from concurrent.futures import ThreadPoolExecutor
from decimal import Decimal, InvalidOperation

from review_quote_sources import QuotesNotPublished

BATCH = 80
FEEDS = {
    "tencent": ("https://qt.gtimg.cn/q=", "https://stock.qq.com/", "腾讯财经"),
    "sina": ("https://hq.sinajs.cn/list=", "https://finance.sina.com.cn/stock/", "新浪财经"),
}
ROWS = re.compile(r'(?:v_|hq_str_)((?:sh|sz)\d{6})="([^"]*)";')
INDICES = ("000001.SH", "399001.SZ", "000300.SH")


def quote_code(code):
    return code[-2:].lower() + code[:6]


def request(url, referer):
    return subprocess.check_output(["curl", "-fsS", "--max-time", "10", "-e", referer, url], stderr=subprocess.PIPE)


def fetch_rows(provider, codes):
    endpoint, referer, label = FEEDS[provider]

    def batch(symbols):
        raw = request(endpoint + ",".join(quote_code(s) for s in symbols), referer)
        expected = {quote_code(s): s for s in symbols}
        result = {}
        for key, payload in ROWS.findall(raw.decode("gb18030")):
            if key not in expected or key in result:
                raise ValueError(f"{label} 返回重复或意外的证券代码")
            result[key] = payload.split("," if provider == "sina" else "~")
        if result.keys() != expected.keys():
            raise ValueError(f"{label} 证券报价覆盖不全")
        return {expected[key]: fields for key, fields in result.items()}

    result = {}
    with ThreadPoolExecutor(max_workers=6) as pool:
        for quotes in pool.map(batch, (codes[i:i+BATCH] for i in range(0, len(codes), BATCH))):
            result.update(quotes)
    return result


def parse_quote(provider, fields, day):
    if provider == "sina":
        stamp = fields[30]
        clock = fields[31].replace(":", "")
        opening, reference, close, high, low = map(Decimal, fields[1:6])
        volume, amount = map(Decimal, fields[8:10])
        status = fields[32]
        if status not in {"00", "03"} or (status == "03" and volume != 0):
            raise ValueError("新浪成交状态异常")
    else:
        stamp = fields[30]
        stamp, clock = f"{stamp[:4]}-{stamp[4:6]}-{stamp[6:8]}", stamp[8:]
        close, reference, opening, volume = map(Decimal, fields[3:7])
        high, low = map(Decimal, fields[33:35])
        amount = Decimal(fields[35].split("/")[2])

    if stamp != day or len(clock) != 6 or not clock.isdecimal() or volume < 0 or amount < 0:
        raise ValueError("行情时点或字段不完整")
    if not volume and not amount and not opening:
        return None
    if not (volume > 0 and amount > 0 and 0 < low <= min(opening, close) <= max(opening, close) <= high
            and reference > 0 and int(clock) >= 150000):
        raise ValueError("行情尚未确认收盘")
    return reference, [day, *map(str, (opening, high, low, close, amount))]


def close_quotes(day, codes):
    errors = []
    for provider, (_, referer, label) in FEEDS.items():
        try:
            benchmarks = fetch_rows(provider, list(INDICES))
            index = {code: parse_quote(provider, fields, day) for code, fields in benchmarks.items()}
            if any(value is None for value in index.values()):
                raise ValueError("主要指数收盘快照未齐")
            data = fetch_rows(provider, codes)
            active, bars = {}, {}
            for code, fields in data.items():
                try:
                    parsed = parse_quote(provider, fields, day)
                except (IndexError, InvalidOperation, ValueError) as exc:
                    raise ValueError(f"{code} 当日收盘报价未齐") from exc
                if parsed:
                    reference, bar = parsed
                    active[code] = reference, Decimal(bar[4])
                    bars[code] = bar
            if not active:
                raise ValueError("没有有效成交报价")
            return active, bars, Decimal(index["000300.SH"][1][4]), referer
        except (OSError, subprocess.CalledProcessError):
            errors.append(f"{label} 报价接口暂不可用")
        except (IndexError, InvalidOperation, UnicodeError, ValueError) as exc:
            errors.append(f"{label} · {exc}")
    raise QuotesNotPublished("；".join(errors))
