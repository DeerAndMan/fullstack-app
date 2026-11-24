"""Market scope and adjusted-price calculations shared by the review APIs."""
import re
from decimal import Decimal

SCOPE = re.compile(r"(?:6\d{5}\.SH|(?:00|30)\d{4}\.SZ)\Z")
VERSION = "shsz-a.high250.ma50.v1"


def number(value):
    return float(round(value, 8))


def update_window(row, quote, factor, window):
    close = Decimal(quote["close"])
    row.update(change=float(quote["pct_chg"]), close=float(close), high=float(quote["high"]),
               factor=float(factor), bars=len(window))
    if len(window) == 250:
        prior = max(list(window)[:-1], key=lambda b: b[1])
        row.update(previousHigh=number(prior[1] / factor), peakDate=prior[0],
                   newHigh=window[-1][1] > prior[1])
    else:
        row["reason"] = "有效日线不足 250 根"
    if row["csi"] and len(window) >= 50:
        total = sum(b[2] for b in list(window)[-50:])
        row.update(ma50=number(total / (50 * factor)), above=50 * close * factor > total)
