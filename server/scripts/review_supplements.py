"""Dated market API evidence used by the daily review update."""
import json
import re
import subprocess
from concurrent.futures import ThreadPoolExecutor
from decimal import Decimal
from fractions import Fraction
from urllib.parse import urlencode

BEIJING_API = "https://api.zizizaizai.com/v3/market/kline/day"


def market_snapshot(stamp):
    url = BEIJING_API + "?" + urlencode(dict(trade_date=stamp, candle_mode=0, limit=8000))
    try:
        raw = subprocess.check_output(["curl", "-fsS", "--compressed", "--retry", "1",
                                       "--max-time", "35", "--retry-max-time", "40", url], stderr=subprocess.PIPE)
        response = json.loads(raw, parse_float=Decimal)
        data = response["data"]
        if response["code"] != 200 or not isinstance(data["list"], list) or data["count"] != len(data["list"]):
            raise ValueError("日线条数不一致")
        return data
    except (subprocess.CalledProcessError, KeyError, TypeError, ValueError) as exc:
        raise ValueError("自在量化市场截面读取失败") from exc


def traded(row):
    if row["is_paused"] not in (0, 1) or row["volume"] < 0:
        raise ValueError("成交状态异常")
    return row["is_paused"] == 0 and row["volume"] > 0


def beijing_window(snapshot, window):
    code, bars = snapshot["ts_code"], window["bars"]
    dates = [r["date"] for r in bars]
    if (not bars or len(bars) > 250 or dates != sorted(set(dates))
            or dates[-1].replace("-", "") != snapshot["trade_date"]
            or any(not 0 < r["close"] <= r["high"] for r in bars)
            or any(bars[-1][key] != snapshot[key] for key in ("close", "high"))):
        raise ValueError(f"{code} 目标日行情与日线不符")
    if len(bars) < 250:
        return False, False
    # Rebase past highs by each ex-rights reference / prior close. Fractions keep an equal high equal.
    events = sorted(window["actions"] or [], key=lambda e: e["Time"])
    scale = Fraction(1)
    today = Fraction(bars[-1]["high"])
    for i in range(len(bars)-2, -1, -1):
        previous = Fraction(bars[i]["close"])
        effective = []
        while events and events[-1]["Time"][:10] > dates[i]:
            effective.append(events.pop())
        reference = previous
        for event in reversed(effective):
            rights = Fraction(event["Peigu"])
            reference = (10*reference - Fraction(event["Fenhong"]) + rights*Fraction(event["Peigujia"])) / (
                10 + rights + Fraction(event["Songzhuangu"]))
        if reference <= 0:
            raise ValueError(f"{code} 除权参考价无效")
        scale *= reference / previous
        if Fraction(bars[i]["high"]) * scale >= today:
            return True, False
    return True, True


def beijing_highs(importer, day, core_codes):
    stamp = day.replace("-", "")
    data = market_snapshot(stamp)
    rows = data["list"]
    symbols = {r["ts_code"] for r in rows}
    if (data.get("trade_date") != stamp or not 0 < len(rows) < 8000 or len(rows) != len(symbols)
            or any(r["trade_date"] != stamp for r in rows) or not core_codes <= symbols):
        raise ValueError("目标交易日的全市场截面尚未齐备")
    listed = [r for r in rows if re.fullmatch(r"\d{6}\.BJ", r["ts_code"])]
    active = [r for r in listed if traded(r)]
    if not active:
        raise ValueError("北交所当日成交截面为空")

    def group(rows):
        codes = [r["ts_code"] for r in rows]
        try:
            raw = subprocess.check_output([str(importer), "-beijing-bars", day, *codes], timeout=100)
        except subprocess.SubprocessError as exc:
            raise ValueError(f"通达信北交所日线或除权数据读取失败（{codes[0]} 等 {len(codes)} 只）") from exc
        windows = json.loads(raw, parse_float=Decimal)
        if set(windows) != set(codes):
            raise ValueError("北交所日线覆盖不全")
        return [(r["ts_code"], *beijing_window(r, windows[r["ts_code"]])) for r in rows]

    hits, eligible, checked = [], 0, 0
    groups = [active[i:i+20] for i in range(0, len(active), 20)]
    with ThreadPoolExecutor(max_workers=3) as pool:
        for batch in pool.map(group, groups):
            for code, enough, hit in batch:
                eligible += enough
                if hit:
                    hits.append(code)
            checked += len(batch)
            if checked % 60 == 0 or checked == len(active):
                print(f"REVIEW 北交所新高已核对 {checked}/{len(active)} 家", flush=True)
    return dict(date=day, source="通达信日线及除权除息接口", rosterSource=BEIJING_API,
                listed=len(listed), traded=len(active), eligible=eligible, highs=len(hits), hits=sorted(hits))


def style_snapshot(importer, day, calendar):
    raw = subprocess.check_output([str(importer), "-style-bars", day], timeout=80)
    series = json.loads(raw)
    observations = []
    for code in ("880003", "880865"):
        bars = series[code]
        dates = [r["date"] for r in bars]
        if len(dates) != 250 or dates != calendar[-len(dates):] or dates[-1] != day:
            raise ValueError(f"{code} 目标日或历史交易日不完整")
        closes = [Decimal(str(r["close"])) for r in bars]
        if any(not x.is_finite() or x <= 0 for x in closes):
            raise ValueError(f"{code} 指数点位异常")

        def ema(n):
            value = sum(closes[:n]) / n
            for close in closes[n:]:
                value += 2 * (close - value) / (n + 1)
            return float(value)

        observations.append(dict(code=code, close=float(closes[-1]), previous=float(closes[-2]),
                                 change=float(100 * (closes[-1] / closes[-2] - 1)), ema5=ema(5), ema13=ema(13)))
    return dict(rows=observations, spread=observations[1]["change"] - observations[0]["change"],
                styleSource="通达信指数日线行情接口", styleDate=day)
